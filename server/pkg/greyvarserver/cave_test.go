package greyvarserver

import (
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

const isle = "isleOfStarting_dev_tiled"

// The debug cave: a cave door in the hill wall at the top left of 1.1 leads
// into a cave holding the teleporter to the debug island, and its mouth
// leads back out below the door.
func TestDebugCaveRoundTrip(t *testing.T) {
	s := tailorServer(t)
	for _, e := range s.entitiesOnGrid(isle, "1.1.tmj") {
		if e.Definition == "teleportPad" {
			t.Fatalf("the debug teleporter should have moved into the cave, found one on 1.1 at %d,%d", e.X, e.Y)
		}
	}

	// The door sits on the hill's edge row; the cell below it is outside.
	var door *Entity
	for _, e := range s.entitiesOnGrid(isle, "1.1.tmj") {
		if e.Definition == "caveDoor" {
			door = e
		}
	}
	if door == nil || door.Y != 0 {
		t.Fatalf("1.1 should have a cave door on its top (hill edge) row: %+v", door)
	}
	below := door.Y + 2*16

	// Standing just below the door does nothing; walking up into it enters.
	rp := playerOn(s, "1.1.tmj", door.X, below)
	s.processEntityInteractions(rp)
	if rp.CurrentGridId != "1.1.tmj" {
		t.Fatal("standing below the cave door must not enter it")
	}
	for i := 0; i < 6 && rp.CurrentGridId == "1.1.tmj"; i++ {
		processMoveRequest(s, rp, &pb.MoveRequest{Y: -1})
	}
	if rp.CurrentGridId != "1.1.debugCave.tmj" || rp.Entity.X != 8*16 || rp.Entity.Y != 12*16 {
		t.Fatalf("walking up into the door should enter the cave at (12,8); on %s at %d,%d", rp.CurrentGridId, rp.Entity.X, rp.Entity.Y)
	}

	// Explore the cave: the teleporter and the way out are reachable, the
	// map's edges (which would lead to 1.2 or 2.1) are not.
	var pad, exit *Entity
	for _, e := range s.entitiesOnGrid(isle, "1.1.debugCave.tmj") {
		switch e.Definition {
		case "teleportPad":
			pad = e
		case "caveExit":
			exit = e
		}
	}
	if pad == nil || exit == nil || pad.Properties["teleport_world"] != "isleOfDebug_dev_tiled" {
		t.Fatalf("cave should hold the debug teleporter and an exit: %v %v", pad, exit)
	}
	type pt struct{ x, y int32 }
	start := pt{rp.Entity.X, rp.Entity.Y}
	seen := map[pt]bool{start: true}
	queue := []pt{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range []pt{{4, 0}, {-4, 0}, {0, 4}, {0, -4}} {
			n := pt{p.x + d.x, p.y + d.y}
			if n.x < 0 || n.y < 0 || n.x > 240 || n.y > 240 || seen[n] || !s.canMoveTo(rp, rp.Entity, n.x, n.y) {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}
	for p := range seen {
		if p.x == 0 || p.y == 0 || p.x == 240 || p.y == 240 {
			t.Fatalf("walked to the cave's edge at (%d,%d)", p.x, p.y)
		}
	}
	if !seen[pt{pad.X, pad.Y}] || !seen[pt{exit.X, exit.Y}] {
		t.Fatal("the teleporter and the exit should both be reachable")
	}

	// Out through the mouth: back below the door, not straight back in.
	rp.Entity.X, rp.Entity.Y = exit.X, exit.Y
	s.processEntityInteractions(rp)
	if rp.CurrentGridId != "1.1.tmj" || rp.Entity.X != door.X || rp.Entity.Y != below {
		t.Fatalf("the cave exit should lead to 1.1 just below the door (%d,%d); on %s at %d,%d", door.X, below, rp.CurrentGridId, rp.Entity.X, rp.Entity.Y)
	}
	s.processEntityInteractions(rp)
	if rp.CurrentGridId != "1.1.tmj" {
		t.Fatal("arriving below the door must not enter the cave again")
	}

	// And the teleporter still takes you to the debug island.
	rp = playerOn(s, "1.1.debugCave.tmj", pad.X, pad.Y)
	s.processEntityInteractions(rp)
	if rp.CurrentWorldId != "isleOfDebug_dev_tiled" {
		t.Fatalf("the cave teleporter should go to the debug island, in %s", rp.CurrentWorldId)
	}
}
