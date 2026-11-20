package greyvarserver

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/world-generator/pkg/puzzlegen"
	log "github.com/sirupsen/logrus"
)

// Ways into rifts (rifts.go): the console `rift` command and the riftPortal
// pad.

func processRiftRequest(s *serverInterface, rp *RemotePlayer, req *pb.RiftRequest) {
	say := func(msg string) { rp.PendingConsoleMessages = append(rp.PendingConsoleMessages, msg) }
	if req.GetLeave() {
		if err := s.leaveRift(rp); err != nil {
			say("rift leave failed: " + err.Error())
			return
		}
		say("left the rift")
		return
	}
	mechanics, err := parseMechanicsList(req.GetMechanics())
	if err != nil {
		say("rift failed: " + err.Error())
		return
	}
	inst, err := s.enterRift(rp, RiftRequest{Level: int(req.GetLevel()), Seed: req.GetSeed(), Mechanics: mechanics})
	if err != nil {
		say("rift failed: " + err.Error())
		return
	}
	say(fmt.Sprintf("entered rift level %d (world %s, seed %d, %d actions at best; best cleared so far: %d). `rift leave` to flee.",
		inst.spec.Level, inst.id, inst.spec.Seed, inst.actions, s.riftLevelCleared(rp.Username)))
}

// parseMechanicsList: nil (the level's own mix) for none.
func parseMechanicsList(raw []string) ([]puzzlegen.Mechanic, error) {
	var names []string
	for _, r := range raw {
		for _, n := range strings.Split(r, ",") {
			if n = strings.TrimSpace(n); n != "" {
				names = append(names, n)
			}
		}
	}
	if len(names) == 0 {
		return nil, nil
	}
	return puzzlegen.ParseMechanics(names)
}

func isRiftPortal(ent *Entity) bool { return ent != nil && ent.Definition == "riftPortal" }

// portalKey names a portal stably (across world reloads): its map position.
func portalKey(ent *Entity) string {
	return fmt.Sprintf("portal:%s/%s/%d,%d", ent.WorldId, ent.GridId, ent.X, ent.Y)
}

// useRiftPortal: a player stepped onto a rift portal. They join the party
// in the rift open from it, or open one at one level above their best (or
// the portal's rift_level).
func (s *serverInterface) useRiftPortal(rp *RemotePlayer, ent *Entity, vis tiled.TypedTileVisual) {
	req := RiftRequest{Key: portalKey(ent)}
	if lvl, err := strconv.Atoi(strings.TrimSpace(entityProp(ent, vis, "rift_level"))); err == nil && lvl > 0 {
		req.Level = lvl
	}
	if raw := entityProp(ent, vis, "rift_mechanics"); raw != "" {
		ms, err := parseMechanicsList([]string{raw})
		if err != nil {
			log.Warnf("riftPortal %d: %v", ent.ServerId, err)
		}
		req.Mechanics = ms
	}
	if _, err := s.enterRift(rp, req); err != nil {
		msg := "The portal fizzles: " + err.Error()
		rp.PendingHudMessage = &msg
		rp.PendingHudEntityId = ent.ServerId
		log.WithFields(log.Fields{"player": rp.Username, "error": err.Error()}).Warn("Rift portal failed")
	}
}

// settleContacts marks what the player is standing on as already touched,
// so arriving on a pad (on a rift's entry teleporter, or back on the portal
// after a rift) does not
// set it off until they step off and on again.
func (s *serverInterface) settleContacts(rp *RemotePlayer) {
	if rp == nil || rp.Entity == nil {
		return
	}
	if rp.ActiveEntityIds == nil {
		rp.ActiveEntityIds = map[int64]bool{}
	}
	for _, ent := range s.entitiesOnGrid(rp.CurrentWorldId, rp.CurrentGridId) {
		if ent.Definition == "player" {
			continue
		}
		vis, ct, ok := s.entityVisualInfo(ent)
		if !ok || ct != tiled.CollisionPressure {
			continue
		}
		if hit, _ := s.entityContactTriggered(rp.Entity, ent, vis, ct); hit {
			rp.ActiveEntityIds[ent.ServerId] = true
		}
	}
}
