package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func TestPlayerMajorityOverlaps(t *testing.T) {
	rects := []tiled.CollisionRect{{X: 0, Y: 0, W: 16, H: 16}}
	if !playerMajorityOverlaps(0, 0, rects, 0, 0) {
		t.Fatal("expected majority on full overlap")
	}
	if playerMajorityOverlaps(-8, 0, rects, 0, 0) {
		t.Fatal("exactly half should not count as majority")
	}
	if !playerMajorityOverlaps(-7, 0, rects, 0, 0) {
		t.Fatal("expected majority when > half overlaps")
	}
	if playerMajorityOverlaps(32, 32, rects, 0, 0) {
		t.Fatal("no overlap should not trigger")
	}
}

func TestTouchAndPressureContact(t *testing.T) {
	grid := &gridfiles.Grid{RowCount: 8, ColCount: 8}
	grid.Build()
	s, rp := testServerWithGrid(grid)
	s.entityDefinitions["signpost"] = &entdefs.EntityDefinition{
		Title: "signpost", InitialState: "signy",
		States: map[string]entdefs.EntityState{"signy": {}},
	}
	s.entityDefinitions["pressureButton"] = &entdefs.EntityDefinition{
		Title: "pressureButton", InitialState: "unpressed",
		States: map[string]entdefs.EntityState{"unpressed": {}, "pressed": {}},
	}
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.entityVisuals.SetTypeForTest("signpost", tiled.TypedTileVisual{
		Frames:        []int32{13},
		Collision:     []tiled.CollisionRect{{X: 0, Y: 0, W: 15, H: 15}},
		CollisionType: tiled.CollisionTouch,
		TileWidth:     15,
		TileHeight:    15,
	})
	s.entityVisuals.SetTypeForTest("pressureButton", tiled.TypedTileVisual{
		Frames:        []int32{20},
		CollisionType: tiled.CollisionPressure,
		TileProps:     map[string]string{"collision_transition": "pressed"},
		TileWidth:     15,
		TileHeight:    15,
	})
	s.entityVisuals.SetTypeForTest("pressureButton_pressed", tiled.TypedTileVisual{
		Frames:        []int32{21},
		CollisionType: tiled.CollisionPressure,
		TileWidth:     15,
		TileHeight:    15,
	})

	player := &Entity{ServerId: 1, Definition: "player", X: 15, Y: 32, WorldId: "testWorld", GridId: "testGrid"}
	sign := &Entity{
		ServerId: 2, Definition: "signpost", State: "signy",
		Properties: map[string]string{"msg": "Hello"},
		X: 32, Y: 32, WorldId: "testWorld", GridId: "testGrid",
	}
	btn := &Entity{
		ServerId: 3, Definition: "pressureButton", State: "unpressed",
		X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid",
	}
	s.entityInstances[1] = player
	s.entityInstances[2] = sign
	s.entityInstances[3] = btn
	rp.Entity = player
	rp.currentFrame = &pb.ServerUpdate{}

	s.processEntityInteractions(rp)
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage != "Hello" {
		t.Fatalf("touch hud = %#v", rp.PendingHudMessage)
	}

	player.X = 0
	player.Y = 0
	rp.PendingHudMessage = nil
	s.processEntityInteractions(rp)
	if btn.State != "pressed" {
		t.Fatalf("pressure state = %q, want pressed", btn.State)
	}

	player.X = 64
	player.Y = 64
	s.processEntityInteractions(rp)
	if btn.State != "unpressed" {
		t.Fatalf("pressure leave state = %q, want unpressed", btn.State)
	}
}

func TestPickupAddsInventoryAndDespawns(t *testing.T) {
	grid := &gridfiles.Grid{RowCount: 8, ColCount: 8}
	grid.Build()
	s, rp := testServerWithGrid(grid)
	s.remotePlayers = map[string]*RemotePlayer{"p": rp}
	s.entityDefinitions["key"] = &entdefs.EntityDefinition{
		Title: "key", InitialState: "keylike",
		States: map[string]entdefs.EntityState{"keylike": {}},
	}
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.entityVisuals.SetTypeForTest("key", tiled.TypedTileVisual{
		Frames:        []int32{0},
		CollisionType: tiled.CollisionPickup,
		TileWidth:     15,
		TileHeight:    15,
	})

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	key := &Entity{ServerId: 2, Definition: "key", State: "keylike", X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	s.entityInstances[2] = key
	rp.Entity = player
	rp.KnownEntities = map[int64]*Entity{2: key}
	rp.currentFrame = &pb.ServerUpdate{}

	s.processEntityInteractions(rp)
	if len(rp.Inventory) != 1 || rp.Inventory[0].Definition != "key" {
		t.Fatalf("inventory = %#v", rp.Inventory)
	}
	if _, ok := s.entityInstances[2]; ok {
		t.Fatal("key should be despawned")
	}
	if len(rp.PendingDespawns) != 1 || rp.PendingDespawns[0] != 2 {
		t.Fatalf("despawns = %#v", rp.PendingDespawns)
	}
}

func TestPressureDoesNotBlockMovement(t *testing.T) {
	grid := &gridfiles.Grid{RowCount: 4, ColCount: 4}
	grid.Build()
	for row := uint32(0); row < 4; row++ {
		for col := uint32(0); col < 4; col++ {
			grid.Tiles[row][col] = &gridfiles.Tile{Row: row, Col: col, Texture: "sand.png", Traversable: true}
		}
	}
	blockingTileTextures = map[string]bool{}
	s, rp := testServerWithGrid(grid)
	s.entityDefinitions["pressureButton"] = &entdefs.EntityDefinition{
		Title: "pressureButton", InitialState: "unpressed", Solid: true,
		States: map[string]entdefs.EntityState{"unpressed": {}},
	}
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.entityVisuals.SetTypeForTest("pressureButton", tiled.TypedTileVisual{
		Frames:        []int32{20},
		Collision:     []tiled.CollisionRect{{X: 0, Y: 0, W: 15, H: 15}},
		CollisionType: tiled.CollisionPressure,
		TileWidth:     15,
		TileHeight:    15,
	})

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	btn := &Entity{ServerId: 2, Definition: "pressureButton", State: "unpressed", X: 16, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	s.entityInstances[2] = btn
	rp.Entity = player

	if blocked, _ := s.isBlockedBySolidEntity(rp, player, 16, 0); blocked {
		t.Fatal("pressure entity must not block movement")
	}
}
