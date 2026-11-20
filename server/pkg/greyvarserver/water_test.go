package greyvarserver

import (
	"path/filepath"
	"testing"

	"github.com/jamesread/greyvar/datlib/gridfiles"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// visibleWater reports whether pixel (x, y) of the grid shows water: the
// topmost terrain tile that is opaque there is water, or water-coloured.
func visibleWater(g *gridfiles.Grid, x, y int32) bool {
	r, c := uint32(y/16), uint32(x/16)
	px, py := uint(x%16), y%16
	for li := len(g.Layers) - 1; li >= 0; li-- {
		if g.Layers[li].Kind != gridfiles.LayerKindTerrain {
			continue
		}
		t := g.Layers[li].TileAt(r, c)
		if t == nil || t.Decor || t.Overlay {
			continue
		}
		if t.SeeThrough != nil && t.SeeThrough[py]&(1<<px) != 0 {
			continue // see-through here: look further down
		}
		return t.Water || (t.WaterPixels != nil && t.WaterPixels[py]&(1<<px) != 0)
	}
	return false
}

// On isleOfStarting the player never wades into water: wherever they can
// walk, (almost) none of their hitbox is over visible water.
func TestPlayerStaysOutOfWater(t *testing.T) {
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
	// The 4px collision grid may leave a sliver of the waterline under the
	// hitbox (8x16 = 128px); wading in would put far more of it in water.
	const maxWaterPx = 24

	for _, grid := range []string{"0.0.tmj", "0.1.tmj", "0.2.tmj", "1.0.tmj", "1.1.tmj", "1.2.tmj", "2.0.tmj", "2.1.tmj"} {
		rp := &RemotePlayer{CurrentWorldId: world, CurrentGridId: grid, KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
		g := s.gridById(world, grid)
		type pos struct{ x, y int32 }
		seen := map[pos]bool{}
		var queue []pos
		for y := int32(0); y <= int32(g.RowCount*16-16); y += 2 {
			for x := int32(0); x <= int32(g.ColCount*16-16); x += 2 {
				if s.isTraversableAt(rp, x, y) {
					p := pos{x, y}
					seen[p] = true
					queue = append(queue, p)
				}
			}
		}
		worst, at := 0, pos{}
		hb := playerHitbox
		for _, p := range queue {
			n := 0
			for y := p.y + hb.Y; y < p.y+hb.Y+hb.H; y++ {
				for x := p.x + hb.X; x < p.x+hb.X+hb.W; x++ {
					if visibleWater(g, x, y) {
						n++
					}
				}
			}
			if n > worst {
				worst, at = n, p
			}
		}
		t.Logf("%s: %d standable spots; most water under hitbox %dpx at (%d,%d)", grid, len(queue), worst, at.x, at.y)
		if worst > maxWaterPx {
			t.Errorf("%s: player can stand with %dpx of water under them at (%d,%d) (cell %d,%d)", grid, worst, at.x, at.y, (at.y+8)/16, (at.x+8)/16)
		}
	}
}

// A see-through shore animates when water shows through it, even past the
// grass edge it is painted over (0.0's west coast), but not where only a
// few water pixels peek through under a sand/grass edge (0.0 inland), nor
// over plain grass (1.1).
func TestShoreAnimatesOverWaterOnly(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	s := newServer()
	const world = "isleOfStarting_dev_tiled"
	still := func(grid string) map[[2]uint32]bool {
		rp := &RemotePlayer{CurrentWorldId: world, CurrentGridId: grid, KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
		g := generateGridUpdate(s, rp)
		mem := s.gridById(world, grid)
		out := map[[2]uint32]bool{}
		for li, l := range g.Layers {
			for _, tl := range l.Tiles {
				if mt := mem.Layers[li].TileAt(tl.Row, tl.Col); mt != nil && mt.NeedsWaterBelow {
					out[[2]uint32{tl.Row, tl.Col}] = tl.Still
				}
			}
		}
		return out
	}
	g00 := still("0.0.tmj")
	for _, rc := range [][2]uint32{{8, 0}, {12, 0}, {1, 3}, {1, 10}, {3, 2}} {
		if s, ok := g00[rc]; !ok || s {
			t.Errorf("0.0 coast %v should animate (shore tile present %v, still %v)", rc, ok, s)
		}
	}
	for _, rc := range [][2]uint32{{8, 1}, {8, 2}, {7, 2}, {2, 5}} {
		if s, ok := g00[rc]; !ok || !s {
			t.Errorf("0.0 inland sand edge %v should stay still (present %v, still %v)", rc, ok, s)
		}
	}
	for rc, s := range still("1.1.tmj") {
		if !s {
			t.Errorf("1.1 %v: sand edge over grass should stay still", rc)
		}
	}
}
