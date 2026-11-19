package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/gridfiles"
)

// Walking into a solid wall from any side must stop with the visible body
// touching it. The player sprite's side-view body is x 4..11 of its 16px
// cell, so the rig hitbox is [4, 0, 8, 16]. With a full-cell box the player
// stopped 4px short when walking left/right.
func TestWallContactIsFlushFromEverySide(t *testing.T) {
	saved := playerHitbox
	playerHitbox = hitbox{4, 0, 8, 16}
	defer func() { playerHitbox = saved }()
	blockingTileTextures = map[string]bool{}

	grid := &gridfiles.Grid{RowCount: 3, ColCount: 3}
	grid.Build()
	for r := uint32(0); r < 3; r++ {
		for c := uint32(0); c < 3; c++ {
			grid.Tiles[r][c] = &gridfiles.Tile{Row: r, Col: c, Texture: "grass.png", Traversable: true, CollisionFromTSX: true}
		}
	}
	grid.Tiles[1][1] = &gridfiles.Tile{Row: 1, Col: 1, Texture: "wall.png", CollisionFromTSX: true,
		Collision: []gridfiles.CollisionRect{{X: 0, Y: 0, W: 16, H: 16}}}

	cases := []struct {
		name           string
		startX, startY int32
		dx, dy         int32
		wantX, wantY   int32
	}{
		{"from left", 0, 16, 1, 0, 4, 16},    // body right edge 4+8 = 12+4 = 16 (wall left)
		{"from right", 32, 16, -1, 0, 28, 16}, // body left edge 28+4 = 32 (wall right)
		{"from above", 16, 0, 0, 1, 16, 0},    // body bottom 0+16 = 16 (wall top)
		{"from below", 16, 32, 0, -1, 16, 32}, // body top 32 (wall bottom)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, rp := testServerWithGrid(grid)
			player := &Entity{ServerId: 1, Definition: "player", X: c.startX, Y: c.startY, WorldId: "testWorld", GridId: "testGrid"}
			s.entityInstances[1] = player
			rp.Entity = player
			for i := 0; i < 20; i++ {
				x, y, ok := s.resolveMoveDestination(rp, player.X+c.dx*DISTANCE_PER_REQUEST, player.Y+c.dy*DISTANCE_PER_REQUEST)
				if !ok {
					break
				}
				player.X, player.Y = x, y
			}
			if player.X != c.wantX || player.Y != c.wantY {
				t.Fatalf("stopped at (%d,%d), want flush at (%d,%d)", player.X, player.Y, c.wantX, c.wantY)
			}
		})
	}
}

// A one-tile doorway: walking straight down while a few pixels off-centre
// slides the player into line (corner sliding) instead of stopping dead.
func TestCornerSlideIntoDoorway(t *testing.T) {
	saved := playerHitbox
	playerHitbox = hitbox{4, 0, 8, 16}
	defer func() { playerHitbox = saved }()
	blockingTileTextures = map[string]bool{}

	grid := &gridfiles.Grid{RowCount: 3, ColCount: 3}
	grid.Build()
	wall := []gridfiles.CollisionRect{{X: 0, Y: 0, W: 16, H: 16}}
	for r := uint32(0); r < 3; r++ {
		for c := uint32(0); c < 3; c++ {
			tile := &gridfiles.Tile{Row: r, Col: c, Texture: "floor.png", Traversable: true, CollisionFromTSX: true}
			if r == 1 && c != 1 { // wall row with a gap at column 1
				tile.Collision = wall
			}
			grid.Tiles[r][c] = tile
		}
	}
	s, rp := testServerWithGrid(grid)
	player := &Entity{ServerId: 1, Definition: "player", X: 22, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	rp.Entity = player

	// Body x 26..34 overlaps the wall at x 32: aligned would be x 12..20.
	for i := 0; i < 20 && player.Y < 32; i++ {
		x, y, ok := s.resolveMoveDestination(rp, player.X, player.Y+DISTANCE_PER_REQUEST)
		if !ok {
			t.Fatalf("stuck at (%d,%d): corner slide should line the player up with the doorway", player.X, player.Y)
		}
		player.X, player.Y = x, y
	}
	if player.Y < 32 {
		t.Fatalf("did not get through the doorway, at (%d,%d)", player.X, player.Y)
	}

	// Far off-centre (no opening within reach) still stops at the wall.
	player.X, player.Y = 0, 0
	if _, _, ok := s.resolveMoveDestination(rp, 0, DISTANCE_PER_REQUEST); ok {
		t.Fatal("no slide when the opening is beyond reach")
	}
}
