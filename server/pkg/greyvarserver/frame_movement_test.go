package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/gridfiles"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func openTestGrid() *gridfiles.Grid {
	g := &gridfiles.Grid{
		ColCount: 16,
		RowCount: 16,
		Tiles:    map[uint32]map[uint32]*gridfiles.Tile{},
	}
	g.BuildEmpty()
	for row := uint32(0); row < 16; row++ {
		for col := uint32(0); col < 16; col++ {
			g.Tiles[row][col] = &gridfiles.Tile{Texture: "sand.png", Traversable: true}
		}
	}
	return g
}

func TestProcessMoveRequestDiagonal(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	s, rp := testServerWithGrid(openTestGrid())
	s.frameTime = 1_000_000_000
	rp.Entity = &Entity{ServerId: 1, Definition: "player", X: 64, Y: 64, GridId: "testGrid"}
	s.entityInstances[1] = rp.Entity

	processMoveRequest(s, rp, &pb.MoveRequest{X: 1, Y: -1})

	if rp.Entity.X != 64+diagonalStep {
		t.Fatalf("x = %d, want %d", rp.Entity.X, 64+diagonalStep)
	}
	if rp.Entity.Y != 64-diagonalStep {
		t.Fatalf("y = %d, want %d", rp.Entity.Y, 64-diagonalStep)
	}
}

func TestResolveMoveDestinationSlidesOnDiagonalBlock(t *testing.T) {
	blockingTileTextures = map[string]bool{"water.png": true}
	grid := openTestGrid()
	// Wall directly south of the player. Diagonal SE is blocked by that wall,
	// but sliding east along open ground should succeed.
	grid.Tiles[2][1] = &gridfiles.Tile{Texture: "water.png"}
	s, rp := testServerWithGrid(grid)
	rp.Entity = &Entity{ServerId: 1, Definition: "player", X: 16, Y: 16, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = rp.Entity

	destX, destY, ok := s.resolveMoveDestination(rp, 20, 20)
	if !ok {
		t.Fatal("expected slide to succeed")
	}
	if destX == 20 && destY == 20 {
		t.Fatal("expected slide, not full diagonal")
	}
	if destX == 16 && destY == 16 {
		t.Fatal("expected movement on at least one axis")
	}
}

func TestResolveMoveDestinationFlushesToContact(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	grid := openTestGrid()
	// Solid collision occupying the full tile at col 2 (x=32..48).
	grid.Tiles[0][2] = &gridfiles.Tile{
		Row:              0,
		Col:              2,
		Texture:          "tree.png",
		CollisionFromTSX: true,
		Collision:        []gridfiles.CollisionRect{{X: 0, Y: 0, W: 16, H: 16}},
	}
	s, rp := testServerWithGrid(grid)
	// Player right edge at x=27; wall left edge at x=32 → 5px gap.
	// Full +6 step would overlap; expect flush at x=16 (right edge == 32).
	rp.Entity = &Entity{ServerId: 1, Definition: "player", X: 11, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = rp.Entity

	destX, destY, ok := s.resolveMoveDestination(rp, 17, 0)
	if !ok {
		t.Fatal("expected partial move to flush contact")
	}
	if destX != 16 || destY != 0 {
		t.Fatalf("dest = (%d,%d), want flush (16,0)", destX, destY)
	}

	// Already flush: further movement into the wall must not move.
	rp.Entity.X = 16
	destX, destY, ok = s.resolveMoveDestination(rp, 22, 0)
	if ok {
		t.Fatalf("expected blocked when already flush, got (%d,%d)", destX, destY)
	}
}

func TestProcessMoveRequestFlushesToContact(t *testing.T) {
	blockingTileTextures = map[string]bool{}
	grid := openTestGrid()
	grid.Tiles[0][2] = &gridfiles.Tile{
		Row:              0,
		Col:              2,
		Texture:          "tree.png",
		CollisionFromTSX: true,
		Collision:        []gridfiles.CollisionRect{{X: 0, Y: 0, W: 16, H: 16}},
	}
	s, rp := testServerWithGrid(grid)
	s.frameTime = 1_000_000_000
	rp.Entity = &Entity{ServerId: 1, Definition: "player", X: 11, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = rp.Entity

	processMoveRequest(s, rp, &pb.MoveRequest{X: 1, Y: 0})

	if rp.Entity.X != 16 || rp.Entity.Y != 0 {
		t.Fatalf("entity = (%d,%d), want flush (16,0)", rp.Entity.X, rp.Entity.Y)
	}
}
