package greyvarserver

import (
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// The big trees on isleOfStarting 1.1 (2x2 tile objects, placed where Tiled
// shows them): only a tree's root blocks; you can walk under the canopy
// (which clients draw over you).
func TestTreeRootBlocksCanopyDoesNot(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	saved := playerHitbox
	defer func() { playerHitbox = saved }()
	s := newServer()
	const world, grid = "isleOfStarting_dev_tiled", "1.1.tmj"
	rp := &RemotePlayer{CurrentWorldId: world, CurrentGridId: grid, KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}

	// Top-left cell of each tree (canopy row, then root row below it).
	trees := [][2]int32{{4, 1}, {3, 2}, {2, 4}}
	want := map[[2]uint32]bool{}
	for _, tr := range trees {
		r, c := tr[0], tr[1]
		for dr := uint32(0); dr < 2; dr++ {
			for dc := uint32(0); dc < 2; dc++ {
				want[[2]uint32{uint32(r) + dr, uint32(c) + dc}] = true
			}
		}
		// The trunk is pixels x+12..x+20, y+24..y+32 of the tree (the
		// player's hitbox is x+4..x+12, y..y+16 of its position).
		x, y := c*16, r*16
		if s.isTraversableAt(rp, x+8, y+16) {
			t.Errorf("tree at (%d,%d): walking into the trunk should be blocked", r, c)
		}
		if !s.isTraversableAt(rp, x+8, y+8) {
			t.Errorf("tree at (%d,%d): standing just behind the trunk, under the canopy, should be allowed", r, c)
		}
	}

	// Their tiles reach the client marked to draw over characters (trees
	// that overlap share a cell).
	got := map[[2]uint32]bool{}
	for _, l := range generateGridUpdate(s, rp).Layers {
		for _, tl := range l.Tiles {
			if tl.Above {
				got[[2]uint32{tl.Row, tl.Col}] = true
				if !want[[2]uint32{tl.Row, tl.Col}] {
					t.Errorf("unexpected above tile at (%d,%d)", tl.Row, tl.Col)
				}
			}
		}
	}
	for cell := range want {
		if !got[cell] {
			t.Errorf("tree tile at (%d,%d) is not marked to draw over characters", cell[0], cell[1])
		}
	}

	// Where two trees overlap (row 4, col 2) both keep their tile, stacked as
	// Tiled draws them: the back tree's root first, then the front tree's
	// canopy over it (grass_terrains tiles 218 and 199).
	var stack []int32
	for _, l := range generateGridUpdate(s, rp).Layers {
		for _, tl := range l.Tiles {
			if tl.Row == 4 && tl.Col == 2 && tl.Above {
				stack = append(stack, tl.FrameIndex)
			}
		}
	}
	if len(stack) != 2 || stack[0] != 218 || stack[1] != 199 {
		t.Fatalf("overlapping trees at (4,2): tiles bottom to top %v, want [218 199]", stack)
	}
}
