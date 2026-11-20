package greyvarserver

import (
	"strings"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func isleRiftPortal(t *testing.T, s *serverInterface) *Entity {
	t.Helper()
	for _, e := range s.entitiesOnGrid("isleOfStarting_dev_tiled", "1.1.tmj") {
		if e.Definition == "riftPortal" {
			return e
		}
	}
	t.Fatal("isleOfStarting 1.1 should have a rift portal")
	return nil
}

func stepOn(s *serverInterface, rp *RemotePlayer, x, y int32) {
	rp.Entity.X, rp.Entity.Y = x, y
	s.processEntityInteractions(rp)
}

// The portal on the spawn grid: step on for a puzzle, friends who follow
// join it, and the exit brings you back onto the portal without sending you
// straight back in.
func TestPuzzlePortalOnIsleOfStarting(t *testing.T) {
	s, ann, bob := partyServer(t)
	portal := isleRiftPortal(t, s)
	if !s.isTraversableOnGrid("isleOfStarting_dev_tiled", "1.1.tmj", portal.X, portal.Y) {
		t.Fatal("the portal should be on open ground")
	}

	stepOn(s, ann, portal.X, portal.Y)
	inst := s.riftFor(ann.CurrentWorldId)
	if inst == nil || inst.spec.Level != 1 {
		t.Fatalf("ann should be in a level 1 puzzle, is in %s", ann.CurrentWorldId)
	}
	stepOn(s, bob, portal.X, portal.Y)
	if bob.CurrentWorldId != inst.id {
		t.Fatalf("bob should join ann's puzzle %s, is in %s", inst.id, bob.CurrentWorldId)
	}

	exit := findRiftEnt(t, s, inst, "exit")
	stepOn(s, bob, exit.X, exit.Y)
	for _, p := range []*RemotePlayer{ann, bob} {
		if p.CurrentWorldId != "isleOfStarting_dev_tiled" || p.Entity.X != portal.X || p.Entity.Y != portal.Y {
			t.Fatalf("%s should be back on the portal, is in %s at %d,%d", p.Username, p.CurrentWorldId, p.Entity.X, p.Entity.Y)
		}
		// Standing (or shuffling) on the portal does not fire it again...
		stepOn(s, p, portal.X+2, portal.Y)
		if s.riftFor(p.CurrentWorldId) != nil {
			t.Fatalf("%s was sent straight back into a puzzle", p.Username)
		}
	}
	// ...stepping off and on again does: the next level.
	stepOn(s, ann, portal.X, portal.Y+32)
	stepOn(s, ann, portal.X, portal.Y)
	next := s.riftFor(ann.CurrentWorldId)
	if next == nil || next.spec.Level != 2 {
		t.Fatalf("stepping back on should start level 2, in %s", ann.CurrentWorldId)
	}
}

func TestPuzzlePortalFixedLevelAndMechanics(t *testing.T) {
	s, ann, _ := partyServer(t)
	portal := isleRiftPortal(t, s)
	portal.Properties = map[string]string{"rift_level": "6", "rift_mechanics": "lever"}
	stepOn(s, ann, portal.X, portal.Y)
	inst := s.riftFor(ann.CurrentWorldId)
	if inst == nil || inst.spec.Level != 6 {
		t.Fatalf("portal puzzle_level 6: in %s", ann.CurrentWorldId)
	}
	for _, e := range inst.result.Puzzle.Ents {
		if typ := e.Kind.Type(); typ == "pushBlock" || typ == "pressureButton" {
			t.Fatalf("lever-only portal placed a %s", typ)
		}
	}
}

func TestConsolePuzzleCommand(t *testing.T) {
	s, ann, _ := partyServer(t)
	last := func() string { return ann.PendingConsoleMessages[len(ann.PendingConsoleMessages)-1] }

	processRiftRequest(s, ann, &pb.RiftRequest{Level: 3, Seed: 4})
	inst := s.riftFor(ann.CurrentWorldId)
	if inst == nil || inst.spec.Level != 3 || inst.spec.Seed != 4 || !strings.Contains(last(), "level 3") {
		t.Fatalf("puzzle 3 seed=4: in %s, said %q", ann.CurrentWorldId, last())
	}
	processRiftRequest(s, ann, &pb.RiftRequest{Level: 2})
	if !strings.Contains(last(), "already in a rift") {
		t.Fatalf("a second puzzle from inside one: %q", last())
	}
	processRiftRequest(s, ann, &pb.RiftRequest{Leave: true})
	if s.riftFor(ann.CurrentWorldId) != nil || s.riftFor(inst.id) != nil || last() != "left the rift" {
		t.Fatalf("leave: in %s, said %q", ann.CurrentWorldId, last())
	}
	processRiftRequest(s, ann, &pb.RiftRequest{Leave: true})
	if !strings.Contains(last(), "not in a rift") {
		t.Fatalf("leave outside a puzzle: %q", last())
	}
	processRiftRequest(s, ann, &pb.RiftRequest{Mechanics: []string{"lever,teleport"}})
	if !strings.Contains(last(), "unknown mechanic") || s.riftFor(ann.CurrentWorldId) != nil {
		t.Fatalf("bad mechanic: %q", last())
	}
	processRiftRequest(s, ann, &pb.RiftRequest{Mechanics: []string{"lever", "keyChest"}})
	if s.riftFor(ann.CurrentWorldId) == nil {
		t.Fatalf("puzzle mechanics=lever,keyChest: %q", last())
	}
}
