package greyvarserver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/world-generator/pkg/puzzlegen"
)

// partyServer is the real server with two players standing on
// isleOfStarting 1.1 (the spawn grid).
func partyServer(t *testing.T) (*serverInterface, *RemotePlayer, *RemotePlayer) {
	t.Helper()
	s := tailorServer(t)
	mk := func(name string, x, y int32) *RemotePlayer {
		ent := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: x, Y: y, WorldId: "isleOfStarting_dev_tiled", GridId: "1.1.tmj"}
		s.entityInstances[ent.ServerId] = ent
		rp := &RemotePlayer{Username: name, Entity: ent, CurrentWorldId: ent.WorldId, CurrentGridId: ent.GridId,
			KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
		rp.key = name
	s.remotePlayers[name] = rp
		return rp
	}
	s.remotePlayers = map[string]*RemotePlayer{}
	return s, mk("ann", 96, 96), mk("bob", 112, 96)
}

func TestPuzzlePartySharesOneInstance(t *testing.T) {
	s, ann, bob := partyServer(t)
	inst, err := s.enterRift(ann, RiftRequest{Level: 2, Seed: 5, Key: "portal1"})
	if err != nil {
		t.Fatal(err)
	}
	if ann.CurrentWorldId != inst.id || s.loadedWorlds[inst.id] == nil {
		t.Fatalf("ann should be in %s, is in %s", inst.id, ann.CurrentWorldId)
	}
	if _, err := os.Stat(filepath.Join(inst.dir, inst.id+".world")); err != nil {
		t.Fatalf("instance files: %v", err)
	}
	if len(s.entitiesOnGrid(inst.id, "0.0.tmj")) < 4 {
		t.Fatal("instance entities should be spawned")
	}

	again, err := s.enterRift(bob, RiftRequest{Key: "portal1"})
	if err != nil || again != inst || bob.CurrentWorldId != inst.id {
		t.Fatalf("bob should join ann's instance: %v %v", again, err)
	}
	if _, err := s.enterRift(bob, RiftRequest{Key: "portal1"}); err == nil {
		t.Fatal("entering a puzzle from inside one should fail")
	}
	// A different key (or none) is a different instance.
	carl := playerOnFor(s, "carl")
	other, err := s.enterRift(carl, RiftRequest{Level: 1, Seed: 1})
	if err != nil || other == inst {
		t.Fatalf("carl should get his own instance: %v", err)
	}

	// Ann leaves: back exactly where she was; the party goes on.
	if err := s.leaveRift(ann); err != nil {
		t.Fatal(err)
	}
	if ann.CurrentWorldId != "isleOfStarting_dev_tiled" || ann.CurrentGridId != "1.1.tmj" || ann.Entity.X != 96 || ann.Entity.Y != 96 {
		t.Fatalf("ann returned to %s/%s %d,%d", ann.CurrentWorldId, ann.CurrentGridId, ann.Entity.X, ann.Entity.Y)
	}
	if s.riftFor(inst.id) == nil {
		t.Fatal("bob is still inside; the instance must stay")
	}

	// Bob disconnects: nobody is left, so it is unmounted and deleted.
	s.removePlayer(bob)
	if s.riftFor(inst.id) != nil || s.loadedWorlds[inst.id] != nil {
		t.Fatal("empty instance should be closed")
	}
	if len(s.entitiesOnGrid(inst.id, "0.0.tmj")) != 0 {
		t.Fatal("instance entities should be purged")
	}
	if _, err := os.Stat(inst.dir); !os.IsNotExist(err) {
		t.Fatalf("instance dir should be deleted: %v", err)
	}
	// The key is free again: the next player there starts a fresh one.
	again, err = s.enterRift(ann, RiftRequest{Level: 1, Seed: 3, Key: "portal1"})
	if err != nil || again == inst {
		t.Fatalf("a new instance should start: %v", err)
	}
}

func playerOnFor(s *serverInterface, name string) *RemotePlayer {
	ent := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 128, Y: 128, WorldId: "isleOfStarting_dev_tiled", GridId: "1.1.tmj"}
	s.entityInstances[ent.ServerId] = ent
	rp := &RemotePlayer{Username: name, Entity: ent, CurrentWorldId: ent.WorldId, CurrentGridId: ent.GridId,
		KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	rp.key = name
	s.remotePlayers[name] = rp
	return rp
}

func TestPuzzleExitRecordsProgressForTheParty(t *testing.T) {
	s, ann, bob := partyServer(t)
	inst, err := s.enterRift(ann, RiftRequest{Level: 1, Seed: 2, Key: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.enterRift(bob, RiftRequest{Key: "p"}); err != nil {
		t.Fatal(err)
	}
	exit := findRiftEnt(t, s, inst, "exit")
	ann.Entity.X, ann.Entity.Y = exit.X, exit.Y
	s.processEntityInteractions(ann)

	for _, p := range []*RemotePlayer{ann, bob} {
		if p.CurrentWorldId != "isleOfStarting_dev_tiled" || p.CurrentGridId != "1.1.tmj" {
			t.Errorf("%s should be back on 1.1, is on %s/%s", p.Username, p.CurrentWorldId, p.CurrentGridId)
		}
		if s.riftLevelCleared(p.Username) != 1 {
			t.Errorf("%s: level solved %d, want 1", p.Username, s.riftLevelCleared(p.Username))
		}
		if p.PendingHudMessage == nil || *p.PendingHudMessage != "Rift level 1 cleared!" {
			t.Errorf("%s: HUD %v", p.Username, p.PendingHudMessage)
		}
	}
	if s.riftFor(inst.id) != nil {
		t.Fatal("instance should close once everyone is out")
	}
	data, err := os.ReadFile(progressPath())
	if err != nil {
		t.Fatal(err)
	}
	saved := map[string]int{}
	if json.Unmarshal(data, &saved); saved["ann"] != 1 || saved["bob"] != 1 {
		t.Fatalf("rifts.json = %s", data)
	}

	// Next time, with no level asked for, ann gets level 2; progress
	// survives a restart.
	next, err := s.enterRift(ann, RiftRequest{})
	if err != nil || next.spec.Level != 2 {
		t.Fatalf("next puzzle: %v %v", next, err)
	}
	s.loadRifts()
	if s.riftLevelCleared("bob") != 1 {
		t.Fatal("progress should be reloaded from rifts.json")
	}
}

func TestStaleInstancesAreRemovedAtStartup(t *testing.T) {
	state := t.TempDir()
	stale := filepath.Join(state, "rifts", "rift_7")
	if err := os.MkdirAll(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	s := tailorServer(t)
	t.Setenv("GREYVAR_STATE_DIR", state)
	s.loadRifts()
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("leftover rift dirs should be removed on start")
	}
}

func TestWorldReloadRegeneratesAnInstance(t *testing.T) {
	s, ann, _ := partyServer(t)
	inst, err := s.enterRift(ann, RiftRequest{Level: 4, Seed: 8})
	if err != nil {
		t.Fatal(err)
	}
	lever := findRiftEnt(t, s, inst, "lever1")
	lever.State = "flickedRight"
	if err := s.reloadWorld(inst.id); err != nil {
		t.Fatal(err)
	}
	if findRiftEnt(t, s, inst, "lever1").State != "flickedLeft" {
		t.Fatal("reload should reset the puzzle")
	}
	if ann.CurrentWorldId != inst.id || s.riftFor(inst.id) == nil {
		t.Fatal("ann should still be in the instance")
	}
}

func findRiftEnt(t *testing.T, s *serverInterface, inst *rift, id string) *Entity {
	t.Helper()
	for _, e := range s.entitiesOnGrid(inst.id, "0.0.tmj") {
		if e.Properties["id"] == id {
			return e
		}
	}
	t.Fatalf("%s: no entity %q", inst.id, id)
	return nil
}

// --- playthrough: the solver's solution, played with the real server rules.

type replay struct {
	t    *testing.T
	s    *serverInterface
	rp   *RemotePlayer
	inst *rift
	// simNow is the simulated clock (ms) for timed levers; each move is one
	// client move interval.
	simNow int64
}

const moveIntervalMs = 67

// tileMs is how long walking one tile takes a real player: 16px at 6px per
// move, one move per client move interval.
const tileMs = 16 * moveIntervalMs / DISTANCE_PER_REQUEST

func (r *replay) cellOf() puzzlegen.Cell {
	return puzzlegen.Cell{R: int((r.rp.Entity.Y + 8) / 16), C: int((r.rp.Entity.X + 8) / 16)}
}

func (r *replay) standable(c puzzlegen.Cell) bool {
	if c.R < 0 || c.C < 0 || c.R >= 16 || c.C >= 16 {
		return false
	}
	// Stepping back onto the entry teleporter would abandon the rift.
	if c == r.inst.result.Puzzle.Spawn {
		return false
	}
	return r.s.canMoveTo(r.rp, r.rp.Entity, int32(c.C*16), int32(c.R*16))
}

// path is a shortest cell path the server lets the player stand along.
func (r *replay) path(to puzzlegen.Cell) []puzzlegen.Cell {
	from := r.cellOf()
	prev := map[puzzlegen.Cell]puzzlegen.Cell{from: from}
	queue := []puzzlegen.Cell{from}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c == to {
			var out []puzzlegen.Cell
			for ; c != from; c = prev[c] {
				out = append([]puzzlegen.Cell{c}, out...)
			}
			return out
		}
		for _, d := range []puzzlegen.Cell{{R: -1}, {R: 1}, {C: -1}, {C: 1}} {
			n := puzzlegen.Cell{R: c.R + d.R, C: c.C + d.C}
			if _, seen := prev[n]; seen || !r.standable(n) {
				continue
			}
			prev[n] = c
			queue = append(queue, n)
		}
	}
	return nil
}

// step walks one cell with real move requests (two 6px moves), then settles
// the last few pixels onto the cell (checked against the same collision).
func (r *replay) step(to puzzlegen.Cell) bool {
	tx, ty := int32(to.C*16), int32(to.R*16)
	dx, dy := sign32(tx-r.rp.Entity.X), sign32(ty-r.rp.Entity.Y)
	for i := 0; i < 2; i++ {
		processMoveRequest(r.s, r.rp, &pb.MoveRequest{X: dx, Y: dy})
	}
	r.simNow += tileMs
	r.s.tickTimers(r.simNow)
	if r.rp.CurrentWorldId != r.inst.id {
		return true // the exit took us away
	}
	if !r.s.canMoveTo(r.rp, r.rp.Entity, tx, ty) {
		return false
	}
	r.rp.Entity.X, r.rp.Entity.Y = tx, ty
	r.s.processEntityInteractions(r.rp)
	return true
}

func (r *replay) tick() {
	r.simNow += moveIntervalMs
	r.s.tickTimers(r.simNow)
}

func (r *replay) walkTo(to puzzlegen.Cell, what string) {
	r.t.Helper()
	if r.cellOf() == to {
		return
	}
	p := r.path(to)
	if p == nil {
		r.t.Fatalf("%s: cannot walk from %v to %v for %s", r.inst.id, r.cellOf(), to, what)
	}
	for _, c := range p {
		if !r.step(c) {
			r.t.Fatalf("%s: blocked stepping to %v for %s (at %d,%d)", r.inst.id, c, what, r.rp.Entity.X, r.rp.Entity.Y)
		}
		if r.rp.CurrentWorldId != r.inst.id {
			return
		}
	}
}

func (r *replay) ent(id string) *Entity {
	r.t.Helper()
	for _, e := range r.s.entitiesOnGrid(r.inst.id, "0.0.tmj") {
		if e.Properties["id"] == id {
			return e
		}
	}
	r.t.Fatalf("%s: no entity %q", r.inst.id, id)
	return nil
}

func (r *replay) do(a puzzlegen.Action) {
	r.t.Helper()
	switch a.Kind {
	case "toggle", "break", "open":
		e := r.ent(a.Entity)
		r.walkTo(a.Stand, a.String())
		if !processEntityInteract(r.s, r.rp, e.ServerId) {
			r.t.Fatalf("%s: %s did nothing", r.inst.id, a)
		}
	case "push":
		e := r.ent(a.Entity)
		r.walkTo(a.Stand, a.String())
		x0, y0 := e.X, e.Y
		for i := 0; i < pushHoldMoves+4 && e.X == x0 && e.Y == y0; i++ {
			processMoveRequest(r.s, r.rp, &pb.MoveRequest{X: int32(a.Dir.C), Y: int32(a.Dir.R)})
			r.tick()
		}
		if e.X != x0+int32(a.Dir.C*16) || e.Y != y0+int32(a.Dir.R*16) {
			r.t.Fatalf("%s: %s: block went from %d,%d to %d,%d", r.inst.id, a, x0, y0, e.X, e.Y)
		}
		// Settle back onto a whole cell.
		r.rp.Entity.X, r.rp.Entity.Y = int32(a.Stand.C*16), int32(a.Stand.R*16)
	case "pull":
		e := r.ent(a.Entity)
		r.walkTo(a.Stand, a.String())
		x0, y0 := e.X, e.Y
		if !processEntityInteract(r.s, r.rp, e.ServerId) {
			r.t.Fatalf("%s: %s: the use button did not pull", r.inst.id, a)
		}
		if e.X != x0+int32(a.Dir.C*16) || e.Y != y0+int32(a.Dir.R*16) || r.cellOf() != a.To {
			r.t.Fatalf("%s: %s: block %d,%d -> %d,%d, player at %v", r.inst.id, a, x0, y0, e.X, e.Y, r.cellOf())
		}
	case "timed":
		e := r.ent(a.Entity)
		r.walkTo(a.Stand, a.String())
		// Run on the simulated clock from the moment of the flip.
		r.simNow = nowMs()
		if !processEntityInteract(r.s, r.rp, e.ServerId) || e.resetAt == 0 {
			r.t.Fatalf("%s: %s: the lever did not start its timer", r.inst.id, a)
		}
		r.walkTo(a.To, a.String())
		r.t.Logf("%s: timed run took %dms of %dms", r.inst.id, r.simNow-(e.resetAt-resetAfterMs(e)), resetAfterMs(e))
	case "exit":
		r.walkTo(a.Stand, a.String())
	}
}

// Generated puzzles can be finished on the real server by following the
// solver's solution: walls, doors, levers, pots, chests, pushes, plates,
// timers and the exit all behave the way the solver assumed.
func TestGeneratedPuzzlesPlayThrough(t *testing.T) {
	levels := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 15}
	seeds := []int64{1, 2, 3}
	if testing.Short() {
		seeds = seeds[:1]
	}
	for _, level := range levels {
		for _, seed := range seeds {
			s, ann, _ := partyServer(t)
			inst, err := s.enterRift(ann, RiftRequest{Level: level, Seed: seed})
			if err != nil {
				t.Fatalf("level %d seed %d: %v", level, seed, err)
			}
			r := &replay{t: t, s: s, rp: ann, inst: inst, simNow: 1}
			for _, a := range inst.result.Solution.Actions {
				r.do(a)
			}
			if ann.CurrentWorldId == inst.id || s.riftLevelCleared("ann") != level {
				t.Fatalf("level %d seed %d: did not finish (in %s, solved %d)\n%s", level, seed, ann.CurrentWorldId,
					s.riftLevelCleared("ann"), puzzlegen.Describe(inst.result))
			}
		}
	}
}

// A block shoved into a corner of a real generated puzzle: the solver's way
// out (a pull) works on the server too.
func TestCorneredBlockReplayUsesPull(t *testing.T) {
	for seed := int64(1); seed <= 3; seed++ {
		s, ann, _ := partyServer(t)
		inst, err := s.enterRift(ann, RiftRequest{Level: 3, Seed: seed})
		if err != nil {
			t.Fatal(err)
		}
		p := inst.result.Puzzle
		var block *puzzlegen.Ent
		for _, e := range p.Ents {
			if e.Kind == puzzlegen.KBlock {
				block = e
			}
		}
		r := p.Rooms[block.Room]
		var corner puzzlegen.Cell
		found := false
		for _, c := range []puzzlegen.Cell{{R: r.R0, C: r.C0}, {R: r.R0, C: r.C1 - 1}, {R: r.R1 - 1, C: r.C0}, {R: r.R1 - 1, C: r.C1 - 1}} {
			taken := c == p.Spawn
			for _, e := range p.Ents {
				taken = taken || (e != block && e.Cell == c)
			}
			if !taken {
				corner, found = c, true
				break
			}
		}
		if !found {
			continue
		}
		block.Cell = corner
		live := findRiftEnt(t, s, inst, block.ID)
		live.X, live.Y = int32(corner.C*16), int32(corner.R*16)

		sol := puzzlegen.Solve(p)
		pulls := 0
		for _, a := range sol.Actions {
			if a.Kind == "pull" {
				pulls++
			}
		}
		if !sol.Solvable || pulls == 0 {
			t.Fatalf("seed %d: cornered block should be solvable by pulling: %v", seed, sol.Actions)
		}
		rr := &replay{t: t, s: s, rp: ann, inst: inst, simNow: 1}
		for _, a := range sol.Actions {
			rr.do(a)
		}
		if s.riftLevelCleared("ann") != 3 {
			t.Fatalf("seed %d: did not finish after %d pulls", seed, pulls)
		}
		return
	}
	t.Fatal("no seed with a free corner")
}

// The player arrives on the entry teleporter without it firing; stepping
// off and back on flees the rift (back where they came from, not cleared),
// while the rest of the party stays in.
func TestRiftEntryTeleporterFlees(t *testing.T) {
	s, ann, bob := partyServer(t)
	inst, err := s.enterRift(ann, RiftRequest{Level: 2, Seed: 6, Key: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.enterRift(bob, RiftRequest{Key: "k"}); err != nil {
		t.Fatal(err)
	}
	entry := findRiftEnt(t, s, inst, "entry")
	if ann.Entity.X != entry.X || ann.Entity.Y != entry.Y {
		t.Fatalf("ann should arrive on the entry teleporter at %d,%d, is at %d,%d", entry.X, entry.Y, ann.Entity.X, ann.Entity.Y)
	}
	s.processEntityInteractions(ann) // standing there does nothing
	if ann.CurrentWorldId != inst.id {
		t.Fatal("arriving on the entry teleporter must not send you straight back")
	}

	sp := inst.result.Puzzle.Spawn
	var off puzzlegen.Cell
	for _, d := range []puzzlegen.Cell{{R: 1}, {R: -1}, {C: 1}, {C: -1}} {
		c := puzzlegen.Cell{R: sp.R + d.R, C: sp.C + d.C}
		if inst.result.Puzzle.Walk[c] {
			off = c
			break
		}
	}
	stepOn(s, ann, int32(off.C*16), int32(off.R*16))
	if ann.CurrentWorldId != inst.id {
		t.Fatal("stepping off the entry teleporter should keep you in the rift")
	}
	stepOn(s, ann, entry.X, entry.Y)
	if ann.CurrentWorldId != "isleOfStarting_dev_tiled" || ann.Entity.X != 96 || ann.Entity.Y != 96 {
		t.Fatalf("stepping back on should flee to where ann came from, in %s at %d,%d", ann.CurrentWorldId, ann.Entity.X, ann.Entity.Y)
	}
	if s.riftLevelCleared("ann") != 0 {
		t.Fatal("fleeing must not count as clearing the rift")
	}
	if ann.PendingHudMessage == nil || *ann.PendingHudMessage != "You fled rift level 2." {
		t.Fatalf("HUD %v", ann.PendingHudMessage)
	}
	if bob.CurrentWorldId != inst.id || s.riftFor(inst.id) == nil {
		t.Fatal("bob is still in the rift; it must stay open")
	}
}

// Rift coins lie by the exit; picking them up adds to a pile that keeps
// growing from rift to rift.
func TestRiftCoinsAccumulate(t *testing.T) {
	s, ann, _ := partyServer(t)
	total := 0
	for i, level := range []int{1, 4} {
		inst, err := s.enterRift(ann, RiftRequest{Level: level, Seed: int64(10 + i)})
		if err != nil {
			t.Fatal(err)
		}
		coins := 0
		for _, e := range s.entitiesOnGrid(inst.id, "0.0.tmj") {
			if e.Definition != "riftCoin" {
				continue
			}
			coins++
			stepOn(s, ann, e.X, e.Y)
		}
		if coins < 2 {
			t.Fatalf("level %d: %d coins", level, coins)
		}
		total += coins
		if got := ann.inventoryCount("rift_coin"); got != int32(total) {
			t.Fatalf("level %d: %d rift coins, want %d", level, got, total)
		}
		if err := s.leaveRift(ann); err != nil {
			t.Fatal(err)
		}
	}
}
