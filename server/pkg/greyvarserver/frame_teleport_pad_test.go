package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/server/pkg/worlds"
)

func TestTeleportPadSameWorld(t *testing.T) {
	src := openGrid(4, 4)
	dst := openGrid(4, 4)
	dst.SpawnPoints = []gridfiles.SpawnPoint{{X: 16, Y: 32}}

	world := &worlds.World{
		ID:        "testWorld",
		SpawnGrid: "src.tmj",
		Grids: map[string]*gridfiles.Grid{
			"src.tmj": src,
			"dst.tmj": dst,
		},
	}

	s, rp := testServerWithGrid(src)
	s.loadedWorlds["testWorld"] = world
	s.entityDefinitions["teleportPad"] = &entdefs.EntityDefinition{
		Title: "teleportPad", InitialState: "idle",
		States: map[string]entdefs.EntityState{"idle": {}},
	}
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.entityVisuals.SetTypeForTest("teleportPad", tiled.TypedTileVisual{
		Frames:        []int32{25},
		CollisionType: tiled.CollisionPressure,
		TileWidth:     15,
		TileHeight:    15,
		TileProps: map[string]string{
			"teleport_world":       "testWorld",
			"teleport_destination": "dst.tmj",
		},
	})

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "testWorld", GridId: "src.tmj"}
	pad := &Entity{
		ServerId: 2, Definition: "teleportPad", State: "idle",
		X: 0, Y: 0, WorldId: "testWorld", GridId: "src.tmj",
	}
	s.entityInstances[1] = player
	s.entityInstances[2] = pad
	rp.Entity = player
	rp.CurrentGridId = "src.tmj"
	rp.KnownEntities = map[int64]*Entity{}
	rp.currentFrame = &pb.ServerUpdate{}

	s.processEntityInteractions(rp)

	if rp.CurrentGridId != "dst.tmj" {
		t.Fatalf("grid = %q, want dst.tmj", rp.CurrentGridId)
	}
	if rp.Entity.X != 16 || rp.Entity.Y != 32 {
		t.Fatalf("pos = (%d,%d), want (16,32)", rp.Entity.X, rp.Entity.Y)
	}
	if !rp.NeedsGridUpdate {
		t.Fatal("expected NeedsGridUpdate")
	}
}

func TestTeleportPadCrossWorld(t *testing.T) {
	src := openGrid(4, 4)
	dst := openGrid(4, 4)
	dst.SpawnPoints = []gridfiles.SpawnPoint{{X: 8, Y: 24}}

	srcWorld := &worlds.World{
		ID: "srcWorld", SpawnGrid: "a.tmj",
		Grids: map[string]*gridfiles.Grid{"a.tmj": src},
	}
	dstWorld := &worlds.World{
		ID: "dstWorld", SpawnGrid: "b.tmj",
		Grids: map[string]*gridfiles.Grid{"b.tmj": dst},
	}

	s, rp := testServerWithGrid(src)
	s.loadedWorlds = map[string]*worlds.World{
		"srcWorld": srcWorld,
		"dstWorld": dstWorld,
	}
	s.remotePlayers = map[string]*RemotePlayer{"p": rp}
	s.entityDefinitions["teleportPad"] = &entdefs.EntityDefinition{
		Title: "teleportPad", InitialState: "idle",
		States: map[string]entdefs.EntityState{"idle": {}},
	}
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.entityVisuals.SetTypeForTest("teleportPad", tiled.TypedTileVisual{
		Frames:        []int32{25},
		CollisionType: tiled.CollisionPressure,
		TileWidth:     15,
		TileHeight:    15,
	})

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "srcWorld", GridId: "a.tmj"}
	pad := &Entity{
		ServerId: 2, Definition: "teleportPad", State: "idle",
		Properties: map[string]string{
			"teleport_world":       "dstWorld",
			"teleport_destination": "b.tmj",
		},
		X: 0, Y: 0, WorldId: "srcWorld", GridId: "a.tmj",
	}
	s.entityInstances[1] = player
	s.entityInstances[2] = pad
	rp.Entity = player
	rp.CurrentWorldId = "srcWorld"
	rp.CurrentGridId = "a.tmj"
	rp.KnownEntities = map[int64]*Entity{1: player, 9: {ServerId: 9}}
	rp.currentFrame = &pb.ServerUpdate{}

	s.processEntityInteractions(rp)

	if rp.CurrentWorldId != "dstWorld" {
		t.Fatalf("world = %q, want dstWorld", rp.CurrentWorldId)
	}
	if rp.CurrentGridId != "b.tmj" {
		t.Fatalf("grid = %q, want b.tmj", rp.CurrentGridId)
	}
	if rp.Entity.X != 8 || rp.Entity.Y != 24 {
		t.Fatalf("pos = (%d,%d), want (8,24)", rp.Entity.X, rp.Entity.Y)
	}
	for _, id := range rp.PendingDespawns {
		if id == player.ServerId {
			t.Fatal("cross-world teleport must not despawn the local player")
		}
	}
}


func TestTeleportPadMissingPropsAndUnknownTargets(t *testing.T) {
	grid := openGrid(4, 4)
	s, rp := testServerWithGrid(grid)
	s.entityDefinitions["teleportPad"] = &entdefs.EntityDefinition{
		Title: "teleportPad", InitialState: "idle",
		States: map[string]entdefs.EntityState{"idle": {}},
	}
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.entityVisuals.SetTypeForTest("teleportPad", tiled.TypedTileVisual{
		Frames:        []int32{25},
		CollisionType: tiled.CollisionPressure,
		TileWidth:     15,
		TileHeight:    15,
	})

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	pad := &Entity{
		ServerId: 2, Definition: "teleportPad", State: "idle",
		X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid",
	}
	s.entityInstances[1] = player
	s.entityInstances[2] = pad
	rp.Entity = player
	rp.currentFrame = &pb.ServerUpdate{}

	s.processEntityInteractions(rp)
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage == "" {
		t.Fatal("expected HUD error for missing teleport_world")
	}
	if rp.CurrentGridId != "testGrid" {
		t.Fatalf("should not move on error, grid=%q", rp.CurrentGridId)
	}

	// Reset contact tracking so enter fires again.
	rp.ActiveEntityIds = nil
	rp.PendingHudMessage = nil
	pad.Properties = map[string]string{
		"teleport_world":       "testWorld",
		"teleport_destination": "",
	}
	s.processEntityInteractions(rp)
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage != "Teleport failed: teleport_destination is missing" {
		t.Fatalf("hud = %#v", rp.PendingHudMessage)
	}

	rp.ActiveEntityIds = nil
	rp.PendingHudMessage = nil
	pad.Properties = map[string]string{
		"teleport_world":       "noSuchWorld_xyz",
		"teleport_destination": "0.0.tmj",
	}
	s.processEntityInteractions(rp)
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage == "" {
		t.Fatal("expected HUD error for missing world")
	}

	rp.ActiveEntityIds = nil
	rp.PendingHudMessage = nil
	pad.Properties = map[string]string{
		"teleport_world":       "testWorld",
		"teleport_destination": "missing.tmj",
	}
	s.processEntityInteractions(rp)
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage != "Teleport failed: destination not found (missing.tmj)" {
		t.Fatalf("hud = %#v", rp.PendingHudMessage)
	}
}

func TestResolveWorldGridAllowsMissingTmjSuffix(t *testing.T) {
	grid := openGrid(2, 2)
	world := &worlds.World{
		Grids: map[string]*gridfiles.Grid{"spawn.tmj": grid},
	}
	id, g := resolveWorldGrid(world, "spawn")
	if g == nil || id != "spawn.tmj" {
		t.Fatalf("got id=%q grid=%v", id, g != nil)
	}
}
