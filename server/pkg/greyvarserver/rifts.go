package greyvarserver

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	common "github.com/jamesread/greyvar/datlib/common"
	"github.com/jamesread/greyvar/datlib/worldfiles"
	"github.com/jamesread/greyvar/server/pkg/worlds"
	"github.com/jamesread/greyvar/world-generator/pkg/puzzlegen"
	log "github.com/sirupsen/logrus"
)

// Rifts are temporary puzzle worlds generated on demand (see
// world-generator/pkg/puzzlegen). One is shared by whoever enters it from
// the same place (a party), lives in $GREYVAR_STATE_DIR/rifts/<id>/ while
// anyone is in it, and is deleted when the last player leaves.
//
//	open     openRift generates, writes, loads and mounts a world
//	enter    enterRift sends a player in, onto the rift's entry teleporter
//	         (remembering where they came from)
//	abandon  stepping back onto the entry teleporter (rift_exit=leave), or
//	         `rift leave`: that player goes back, the rift stays failed for
//	         them (the rest of the party carries on)
//	complete the exit teleporter (rift_exit=complete) records the level as
//	         cleared for everyone in the rift and sends them back
//	close    playerLeftWorld: nobody left in it -> unmount and delete
//
// Rift coins lie around the exit (pickups, item rift_coin); they are a
// reward for getting that far and simply accumulate in the inventory.
//
// Progress (each player's highest level cleared) is kept in
// $GREYVAR_STATE_DIR/rifts.json.

const riftPrefix = "rift_"

// riftOrigin is where a player entered a rift from.
type riftOrigin struct {
	World, Grid string
	X, Y        int32
}

type rift struct {
	id   string
	spec puzzlegen.Spec
	dir  string
	// key is what started it (e.g. a portal), so others there join the
	// same instance; "" for a private one.
	key     string
	origins map[string]riftOrigin // by username
	actions int                     // the solver's shortest solution
	result  *puzzlegen.Result       // the verified layout and its solution
}

type riftState struct {
	instances map[string]*rift
	byKey     map[string]string // key -> instance id
	next      int
	progress  map[string]int // username -> highest level solved
	rng       *rand.Rand
}

func stateDir() string { return cropStateDir() }

func riftsDir() string { return filepath.Join(stateDir(), "rifts") }

func progressPath() string { return filepath.Join(stateDir(), "rifts.json") }

// loadRifts resets rift state: stale rift dirs from a previous run
// are removed and saved progress is read.
func (s *serverInterface) loadRifts() {
	s.rifts = &riftState{
		instances: map[string]*rift{},
		byKey:     map[string]string{},
		progress:  map[string]int{},
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	if err := os.RemoveAll(riftsDir()); err != nil {
		log.Warnf("Cannot remove stale rifts: %v", err)
	}
	data, err := os.ReadFile(progressPath())
	if err == nil {
		if err := json.Unmarshal(data, &s.rifts.progress); err != nil {
			log.Warnf("Cannot read %s: %v", progressPath(), err)
		}
	}
}

func (s *serverInterface) riftState() *riftState {
	if s.rifts == nil {
		s.loadRifts()
	}
	return s.rifts
}

func (s *serverInterface) saveProgress() {
	ps := s.riftState()
	data, err := json.MarshalIndent(ps.progress, "", "  ")
	if err == nil {
		err = os.MkdirAll(stateDir(), 0o755)
	}
	if err == nil {
		tmp := progressPath() + ".tmp"
		if err = os.WriteFile(tmp, data, 0o644); err == nil {
			err = os.Rename(tmp, progressPath())
		}
	}
	if err != nil {
		log.Warnf("Cannot save rift progress: %v", err)
	}
}

// riftLevelCleared is the highest rift level the player has cleared.
func (s *serverInterface) riftLevelCleared(username string) int {
	return s.riftState().progress[username]
}

// riftFor is the rift a world id belongs to, or nil.
func (s *serverInterface) riftFor(worldId string) *rift {
	if s.rifts == nil {
		return nil
	}
	return s.rifts.instances[worldId]
}

// generateRift writes a verified puzzle rift for spec into dir and loads it.
func generateRift(id, dir string, spec puzzlegen.Spec) (*worlds.World, *puzzlegen.Result, error) {
	if err := os.RemoveAll(dir); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	tilesets := filepath.Join(common.ResDir(), "img", "textures", "tilesets")
	res, err := puzzlegen.Generate(spec, puzzlegen.Options{
		WorldName:        id,
		OutDir:           dir,
		TilesetPath:      filepath.Join(tilesets, "grass_terrains.tsx"),
		DecorTilesetPath: filepath.Join(tilesets, "flora.tsx"),
		EntdefsDir:       filepath.Join(common.DatDir(), "entdefs"),
	})
	if err != nil {
		return nil, nil, err
	}
	world, err := worldfiles.LoadWorldFromDir(id, dir)
	if err != nil {
		return nil, nil, err
	}
	world.ID = id
	return world, res, nil
}

// mountRiftWorld makes a loaded rift world live: entities spawned, doors set.
func (s *serverInterface) mountRiftWorld(world *worlds.World) {
	s.loadedWorlds[world.ID] = world
	for gridId, grid := range world.Grids {
		s.instantiateGridEntities(world.ID, gridId, grid)
		s.evaluateOpenWhen(world.ID, gridId, 0)
	}
}

// openRift generates and mounts a new rift (nobody is in it yet).
func (s *serverInterface) openRift(spec puzzlegen.Spec, key string) (*rift, error) {
	ps := s.riftState()
	ps.next++
	id := fmt.Sprintf("%s%d", riftPrefix, ps.next)
	for s.loadedWorlds[id] != nil {
		ps.next++
		id = fmt.Sprintf("%s%d", riftPrefix, ps.next)
	}
	dir := filepath.Join(riftsDir(), id)
	t0 := time.Now()
	world, res, err := generateRift(id, dir, spec)
	if err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}
	inst := &rift{id: id, spec: spec, dir: dir, key: key, origins: map[string]riftOrigin{}, actions: len(res.Solution.Actions) - 1, result: res}
	ps.instances[id] = inst
	if key != "" {
		ps.byKey[key] = id
	}
	s.mountRiftWorld(world)
	log.WithFields(log.Fields{
		"world": id, "level": spec.Level, "seed": spec.Seed, "actions": inst.actions,
		"took": time.Since(t0).Round(time.Millisecond),
	}).Info("Opened rift")
	return inst, nil
}

// RiftRequest says which rift (puzzle level) a player wants.
type RiftRequest struct {
	Level     int    // 0: one more than the player's best
	Seed      int64  // 0: random
	Mechanics []puzzlegen.Mechanic
	// Key shares the instance: everyone entering with the same key while it
	// runs joins it (e.g. a portal's id). "" is a private instance.
	Key string
}

// enterRift joins the running instance for req.Key, or starts a new one,
// and sends rp into it. It returns the instance.
func (s *serverInterface) enterRift(rp *RemotePlayer, req RiftRequest) (*rift, error) {
	if rp == nil || rp.Entity == nil {
		return nil, fmt.Errorf("no player")
	}
	if s.riftFor(rp.CurrentWorldId) != nil {
		return nil, fmt.Errorf("already in a rift (leave it first)")
	}
	ps := s.riftState()
	var inst *rift
	if req.Key != "" {
		inst = ps.instances[ps.byKey[req.Key]]
	}
	if inst == nil {
		level := req.Level
		if level <= 0 {
			level = s.riftLevelCleared(rp.Username) + 1
		}
		seed := req.Seed
		if seed == 0 {
			seed = ps.rng.Int63n(1<<31) + 1
		}
		var err error
		inst, err = s.openRift(puzzlegen.Curve(level, seed, req.Mechanics), req.Key)
		if err != nil {
			return nil, err
		}
	}
	inst.origins[rp.Username] = riftOrigin{World: rp.CurrentWorldId, Grid: rp.CurrentGridId, X: rp.Entity.X, Y: rp.Entity.Y}
	if err := s.loadPlayerWorld(rp, inst.id); err != nil {
		delete(inst.origins, rp.Username)
		s.playerLeftWorld(inst.id)
		return nil, err
	}
	// They arrive on the entry teleporter: it only fires once they have
	// stepped off it and come back.
	s.settleContacts(rp)
	msg := fmt.Sprintf("Rift level %d. Reach the exit teleporter, or step back on this one to flee.", inst.spec.Level)
	if len(s.playersInWorld(inst.id)) > 1 {
		msg = fmt.Sprintf("Joined the party in rift level %d.", inst.spec.Level)
	}
	rp.PendingHudMessage = &msg
	rp.PendingHudEntityId = 0
	return inst, nil
}

func (s *serverInterface) playersInWorld(worldId string) []*RemotePlayer {
	var out []*RemotePlayer
	for _, p := range s.remotePlayers {
		if p.CurrentWorldId == worldId && p.Entity != nil {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Username < out[j].Username })
	return out
}

// returnFromRift sends rp back to where they entered inst from (or the
// default world's spawn).
func (s *serverInterface) returnFromRift(rp *RemotePlayer, inst *rift) error {
	origin, ok := inst.origins[rp.Username]
	delete(inst.origins, rp.Username)
	var err error
	if !ok || s.riftFor(origin.World) != nil {
		err = s.loadPlayerWorld(rp, s.defaultWorldId)
	} else {
		err = s.placePlayer(rp, origin.World, origin.Grid, origin.X, origin.Y)
	}
	// Back on the portal: it must not fire again until they step off it.
	s.settleContacts(rp)
	return err
}

// leaveRift takes rp out of the rift they are in, failed.
func (s *serverInterface) leaveRift(rp *RemotePlayer) error {
	inst := s.riftFor(rp.CurrentWorldId)
	if inst == nil {
		return fmt.Errorf("not in a rift")
	}
	if err := s.returnFromRift(rp, inst); err != nil {
		return err
	}
	msg := fmt.Sprintf("You fled rift level %d.", inst.spec.Level)
	rp.PendingHudMessage = &msg
	rp.PendingHudEntityId = 0
	return nil
}

// completeRift: someone reached the exit. Everyone in the rift has
// cleared it; each goes back to where they came from.
func (s *serverInterface) completeRift(rp *RemotePlayer) {
	inst := s.riftFor(rp.CurrentWorldId)
	if inst == nil {
		msg := "This exit only works in a rift."
		rp.PendingHudMessage = &msg
		return
	}
	ps := s.riftState()
	level := inst.spec.Level
	for _, p := range s.playersInWorld(inst.id) {
		if level > ps.progress[p.Username] {
			ps.progress[p.Username] = level
		}
		if err := s.returnFromRift(p, inst); err != nil {
			log.Warnf("Cannot return %s from %s: %v", p.Username, inst.id, err)
			continue
		}
		msg := fmt.Sprintf("Rift level %d cleared!", level)
		p.PendingHudMessage = &msg
		p.PendingHudEntityId = 0
	}
	s.saveProgress()
	log.WithFields(log.Fields{"world": inst.id, "level": level, "by": rp.Username}).Info("Rift cleared")
}

// placePlayer moves rp to a position in any loaded (or loadable) world.
func (s *serverInterface) placePlayer(rp *RemotePlayer, worldId, gridId string, x, y int32) error {
	world, err := s.ensureWorldLoaded(worldId)
	if err != nil {
		return err
	}
	if world.Grids[gridId] == nil {
		return s.loadPlayerWorld(rp, worldId)
	}
	old := rp.CurrentWorldId
	s.leaveCurrentLocation(rp)
	rp.PendingGridTransition = nil
	rp.CurrentWorldId, rp.CurrentGridId = worldId, gridId
	rp.Entity.WorldId, rp.Entity.GridId = worldId, gridId
	rp.Entity.X, rp.Entity.Y = x, y
	rp.NeedsGridUpdate = true
	if old != worldId {
		s.playerLeftWorld(old)
	}
	return nil
}

// playerLeftWorld is called after a player has left worldId (another world,
// or the game). A rift nobody is in any more is closed.
func (s *serverInterface) playerLeftWorld(worldId string) {
	inst := s.riftFor(worldId)
	if inst == nil || len(s.playersInWorld(worldId)) > 0 {
		return
	}
	s.closeRift(inst)
}

// closeRift unmounts a rift and deletes its files.
func (s *serverInterface) closeRift(inst *rift) {
	ps := s.riftState()
	s.purgeWorldEntities(inst.id)
	delete(s.loadedWorlds, inst.id)
	delete(ps.instances, inst.id)
	if inst.key != "" && ps.byKey[inst.key] == inst.id {
		delete(ps.byKey, inst.key)
	}
	if err := os.RemoveAll(inst.dir); err != nil {
		log.Warnf("Cannot remove %s: %v", inst.dir, err)
	}
	log.WithFields(log.Fields{"world": inst.id}).Info("Closed rift")
}

// regenerateRift rebuilds a rift's world from its spec (for
// world-reload), keeping who is in it.
func (s *serverInterface) regenerateRift(inst *rift) error {
	s.purgeWorldEntities(inst.id)
	delete(s.loadedWorlds, inst.id)
	world, res, err := generateRift(inst.id, inst.dir, inst.spec)
	if err != nil {
		return err
	}
	inst.actions = len(res.Solution.Actions) - 1
	inst.result = res
	s.mountRiftWorld(world)
	return nil
}
