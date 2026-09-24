package greyvarserver

import (
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/server/pkg/worlds"
)

func openGrid(rows, cols uint32) *gridfiles.Grid {
	grid := &gridfiles.Grid{
		RowCount: rows,
		ColCount: cols,
	}
	grid.Build()
	for _, pos := range grid.CellIterator() {
		grid.Tiles[pos.Row][pos.Col] = &gridfiles.Tile{
			Row:         pos.Row,
			Col:         pos.Col,
			Texture:     "grass.png",
			Traversable: true,
		}
	}
	return grid
}

func testTransitionWorld() (*serverInterface, *RemotePlayer, *worlds.World) {
	grid00 := openGrid(4, 4)
	grid01 := openGrid(4, 4)
	grid10 := openGrid(4, 4)
	grid20 := openGrid(4, 4)

	world := &worlds.World{
		ID:        "testWorld",
		SpawnGrid: "0.0.tmj",
		Grids: map[string]*gridfiles.Grid{
			"0.0.tmj": grid00,
			"0.1.tmj": grid01,
			"1.0.tmj": grid10,
			"2.0.tmj": grid20,
		},
	}

	player := &Entity{
		ServerId:   1,
		Definition: "player",
		X:          0,
		Y:          0,
		WorldId:    "testWorld",
		GridId:     "1.0.tmj",
	}

	s := &serverInterface{
		loadedWorlds:    map[string]*worlds.World{"testWorld": world},
		entityInstances: map[int64]*Entity{player.ServerId: player},
		remotePlayers:   map[string]*RemotePlayer{},
	}

	rp := &RemotePlayer{
		Username:       "bob",
		CurrentWorldId: "testWorld",
		CurrentGridId:  "1.0.tmj",
		Entity:         player,
		KnownEntities:  map[int64]*Entity{2: {ServerId: 2}},
	}

	return s, rp, world
}

func TestEntityWithinGridBounds(t *testing.T) {
	grid := openGrid(4, 4)
	if !entityWithinGridBounds(grid, 0, 0) {
		t.Fatal("expected origin in bounds")
	}
	if !entityWithinGridBounds(grid, 48, 48) {
		t.Fatal("expected bottom-right entity origin in bounds")
	}
	if entityWithinGridBounds(grid, -4, 0) {
		t.Fatal("expected negative x out of bounds")
	}
	if entityWithinGridBounds(grid, 0, -4) {
		t.Fatal("expected negative y out of bounds")
	}
	if entityWithinGridBounds(grid, 49, 0) {
		t.Fatal("expected east overflow out of bounds")
	}
}

func TestCanMoveToNorthAndWestEdges(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	s, rp, _ := testTransitionWorld()

	if s.canMoveTo(rp, rp.Entity, 0, -4) {
		t.Fatal("expected move north off grid to be blocked")
	}
	if s.canMoveTo(rp, rp.Entity, -4, 0) {
		t.Fatal("expected move west off grid to be blocked")
	}
}

func TestTryEdgeTransitionNorthAndWest(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	s, rp, world := testTransitionWorld()

	if !s.tryEdgeTransition(rp, &pb.MoveRequest{Y: -1}, 0, -4) {
		t.Fatal("expected north edge transition from 1.0 to 0.0")
	}
	if rp.CurrentGridId != "0.0.tmj" {
		t.Fatalf("grid = %q, want 0.0.tmj", rp.CurrentGridId)
	}
	_, _, _, maxY := gridPixelBounds(world.Grids["0.0.tmj"])
	if rp.Entity.Y != maxY {
		t.Fatalf("entry y = %d, want %d", rp.Entity.Y, maxY)
	}

	rp.CurrentGridId = "0.1.tmj"
	rp.Entity.GridId = "0.1.tmj"
	rp.Entity.X = 0
	rp.Entity.Y = 0

	if !s.tryEdgeTransition(rp, &pb.MoveRequest{X: -1}, -4, 0) {
		t.Fatal("expected west edge transition from 0.1 to 0.0")
	}
	if rp.CurrentGridId != "0.0.tmj" {
		t.Fatalf("grid = %q, want 0.0.tmj", rp.CurrentGridId)
	}
}

func TestProcessMoveRequestTriggersNorthTransition(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	s, rp, _ := testTransitionWorld()
	s.frameTime = 1_000_000_000

	processMoveRequest(s, rp, &pb.MoveRequest{Y: -1})

	if rp.CurrentGridId != "0.0.tmj" {
		t.Fatalf("grid = %q, want 0.0.tmj after north move", rp.CurrentGridId)
	}
}

func TestTryEdgeTransitionNearSouthEdge(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	s, rp, _ := testTransitionWorld()
	_, _, _, maxY := gridPixelBounds(s.gridForPlayer(rp))

	rp.CurrentGridId = "1.0.tmj"
	rp.Entity.GridId = "1.0.tmj"
	rp.Entity.X = 16
	rp.Entity.Y = maxY - 2

	if !s.tryEdgeTransition(rp, &pb.MoveRequest{Y: 1}, rp.Entity.X, rp.Entity.Y+4) {
		t.Fatal("expected south transition when blocked move would leave grid")
	}
	if rp.CurrentGridId != "2.0.tmj" {
		t.Fatalf("grid = %q, want 2.0.tmj", rp.CurrentGridId)
	}
}

func TestTransitionDoesNotDespawnSelf(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	s, rp, _ := testTransitionWorld()
	selfID := rp.Entity.ServerId
	rp.KnownEntities[selfID] = rp.Entity

	s.tryEdgeTransition(rp, &pb.MoveRequest{Y: -1}, 0, -4)

	for _, id := range rp.PendingDespawns {
		if id == selfID {
			t.Fatal("transition should not despawn the moving player")
		}
	}
}
