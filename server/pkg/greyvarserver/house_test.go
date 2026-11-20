package greyvarserver

import (
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// isleOfStarting 1.2: walk up into the red house's door, arrive inside, walk
// down through the doorway, arrive back in front of the house.
func TestRedHouseDoorRoundTrip(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	saved := playerHitbox
	defer func() { playerHitbox = saved }()

	s := newServer()
	const world = "isleOfStarting_dev_tiled"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 4 * 16, Y: 7 * 16, WorldId: world, GridId: "1.2.tmj", State: "idle"}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "guest", Entity: player, CurrentWorldId: world, CurrentGridId: "1.2.tmj",
		KnownEntities: map[int64]*Entity{}, KnownEntdefs: map[string]bool{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"guest": rp}

	// The house is solid; only its door lets you in.
	for _, col := range []int32{3, 5, 6} {
		if s.isTraversableAt(rp, col*16, 5*16) {
			t.Errorf("house wall at (5,%d) should block", col)
		}
	}

	// walk holds a direction until the player changes grid (or gives up).
	walk := func(dy int32, steps int) {
		from := rp.CurrentGridId
		for i := 0; i < steps && rp.CurrentGridId == from; i++ {
			s.frameTime += 70_000_000
			processMoveRequest(s, rp, &pb.MoveRequest{X: 0, Y: dy})
		}
	}

	walk(-1, 12) // up the path into the door
	if rp.CurrentGridId != "1.2.redHouse.tmj" {
		t.Fatalf("walking into the door should go inside, still on %s at (%d,%d)", rp.CurrentGridId, player.X, player.Y)
	}
	if rp.PendingGridTransition != nil {
		t.Fatalf("a door is a warp (fade through black), not a scroll: %+v", rp.PendingGridTransition)
	}
	if player.X != 7*16 || player.Y != 9*16 {
		t.Fatalf("should arrive just inside the doorway (7,9), got (%d,%d)", player.X/16, player.Y/16)
	}

	walk(+1, 12) // back down through the doorway
	if rp.CurrentGridId != "1.2.tmj" {
		t.Fatalf("walking out of the doorway should leave the house, still on %s at (%d,%d)", rp.CurrentGridId, player.X, player.Y)
	}
	if rp.PendingGridTransition != nil {
		t.Fatal("leaving through the doorway should fade too, not scroll")
	}
	if player.X != 4*16 || player.Y != 6*16 {
		t.Fatalf("should arrive in front of the door (4,6), got (%d,%d)", player.X/16, player.Y/16)
	}
}

// Doors and pads are walked onto, so clients draw them beneath characters.
func TestWalkOnEntitiesAreFloor(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	s := newServer()
	for _, def := range []string{"houseDoor", "stoneDoorway", "teleportPad", "chest", "bill", "player"} {
		s.loadEntdef(def)
	}
	rp := &RemotePlayer{KnownEntdefs: map[string]bool{}, currentFrame: &pb.ServerUpdate{}}
	frameNewEntdefs(s, rp)
	floor := map[string]bool{}
	for _, d := range rp.currentFrame.EntityDefinitions {
		floor[d.Name] = d.Floor
	}
	for def, want := range map[string]bool{"houseDoor": true, "stoneDoorway": true, "teleportPad": true, "chest": false, "bill": false, "player": false} {
		if floor[def] != want {
			t.Errorf("%s: floor = %v, want %v", def, floor[def], want)
		}
	}
}

// Every item reaches the client with its HUD icon.
func TestItemIconsAreSent(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	s := newServer()
	s.loadItemCatalog()
	rp := &RemotePlayer{currentFrame: &pb.ServerUpdate{}}
	frameItemDefinitions(s, rp)
	if len(rp.currentFrame.ItemDefinitions) == 0 {
		t.Fatal("no item definitions sent")
	}
	for _, d := range rp.currentFrame.ItemDefinitions {
		if d.Icon != "img/textures/items/icons/"+d.Id+".png" {
			t.Errorf("%s: icon %q", d.Id, d.Icon)
		}
	}
}
