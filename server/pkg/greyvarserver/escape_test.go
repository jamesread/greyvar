package greyvarserver

import (
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// Plays the real cellOfEscape map (real entdefs, items and map data):
// pot hides the Rusty Key -> key opens the chest (note) -> levers
// RIGHT, LEFT, RIGHT open the cell door -> the escape pad warps to the island.
func TestCellOfEscapeIsSolvable(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "cellOfEscape_dev_tiled")
	saved := playerHitbox
	defer func() { playerHitbox = saved }()

	s := newServer()
	const world, grid = "cellOfEscape_dev_tiled", "0.0.tmj"
	find := func(def string, row, col int32) *Entity {
		for _, e := range s.entitiesOnGrid(world, grid) {
			if e.Definition == def && e.X == col*16 && e.Y == row*16 {
				return e
			}
		}
		t.Fatalf("no %s at (%d,%d)", def, row, col)
		return nil
	}
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 4 * 16, Y: 6 * 16, WorldId: world, GridId: grid, State: "idle"}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "escaper", Entity: player, CurrentWorldId: world, CurrentGridId: grid,
		KnownEntities: map[int64]*Entity{}, KnownEntdefs: map[string]bool{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"escaper": rp}

	// reachable: a search in 2px steps (like real movement, which is pixel by
	// pixel) over positions the server lets the player occupy.
	reachable := func(toRow, toCol int32) bool {
		type pos struct{ x, y int32 }
		start := pos{player.X, player.Y}
		goal := pos{toCol * 16, toRow * 16}
		seen := map[pos]bool{start: true}
		queue := []pos{start}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			if cur == goal {
				return true
			}
			for _, d := range []pos{{2, 0}, {-2, 0}, {0, 2}, {0, -2}} {
				n := pos{cur.x + d.x, cur.y + d.y}
				if seen[n] {
					continue
				}
				if !s.isTraversableAt(rp, n.x, n.y) {
					continue
				}
				if blocked, _ := s.isBlockedBySolidEntity(rp, player, n.x, n.y); blocked {
					continue
				}
				seen[n] = true
				queue = append(queue, n)
			}
		}
		return false
	}
	standAt := func(row, col int32) {
		if !reachable(row, col) {
			t.Fatalf("cannot walk to (%d,%d)", row, col)
		}
		player.X, player.Y = col*16, row*16
	}
	interact := func(e *Entity) {
		if !processEntityInteract(s, rp, e.ServerId) {
			t.Fatalf("interacting with %s at (%d,%d) did nothing", e.Definition, e.Y/16, e.X/16)
		}
	}

	door := find("cellDoor", 8, 13)
	if door.State != "closed" || reachable(10, 13) {
		t.Fatal("the escape pad must be shut in by the cell door at the start")
	}

	chest := find("chest", 10, 8)
	standAt(10, 7)
	processEntityInteract(s, rp, chest.ServerId)
	if chest.State != "closed" {
		t.Fatal("the chest should be locked without the key")
	}

	standAt(9, 3)
	interact(find("pot", 9, 2))
	if rp.inventoryCount("rusty_key") != 1 {
		t.Fatal("the pot at (9,2) should hide the Rusty Key")
	}

	standAt(10, 7)
	interact(chest)
	if chest.State != "open" || rp.inventoryCount("note") != 1 || rp.inventoryCount("rusty_key") != 0 {
		t.Fatal("the key should open the chest and give the note")
	}

	standAt(4, 10)
	interact(find("wallSwitch", 3, 10)) // sw1 -> right
	standAt(4, 13)
	interact(find("wallSwitch", 3, 13)) // sw3 -> right (sw2 stays left)
	if door.State != "open" {
		t.Fatalf("RIGHT, LEFT, RIGHT should open the cell door (states: %s %s %s)",
			find("wallSwitch", 3, 10).State, find("wallSwitch", 3, 11).State, find("wallSwitch", 3, 13).State)
	}

	standAt(10, 13)
	s.processEntityInteractions(rp)
	if rp.CurrentWorldId != "isleOfStarting_dev_tiled" {
		t.Fatalf("stepping on the pad should escape to the island, in %s/%s", rp.CurrentWorldId, rp.CurrentGridId)
	}
}
