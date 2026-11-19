package greyvarserver

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jamesread/greyvar/datlib/gridfiles"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// hillTiles are grass_terrains tiles with a "high" (hill) corner.
var hillTiles = map[int32]bool{}

func init() {
	for _, id := range []int32{10, 11, 30, 31, 49, 50, 51, 69, 70, 71, 89, 90, 91,
		129, 130, 131, 149, 151, 169, 170, 171, 192, 193, 194, 212, 214, 232, 233, 234,
		249, 250, 251, 269, 271, 289, 290, 291} {
		hillTiles[id] = true
	}
}

// baseTile is the cell's terrain tile (not decor or a generated overlay).
func baseTile(g *gridfiles.Grid, r, c uint32) *gridfiles.Tile {
	for li := len(g.Layers) - 1; li >= 0; li-- {
		if tl := g.Layers[li].TileAt(r, c); tl != nil && !tl.Decor && !tl.Overlay {
			return tl
		}
	}
	return nil
}

// Every hill on isleOfStarting is walled in by its collision: from low
// ground the only way onto a plateau is a staircase.
func TestHillsAreSealed(t *testing.T) {
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

	for _, grid := range []string{"0.0.tmj", "0.1.tmj", "0.2.tmj", "1.0.tmj", "1.1.tmj", "1.2.tmj"} {
		rp := &RemotePlayer{CurrentWorldId: world, CurrentGridId: grid, KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
		g := s.gridById(world, grid)
		cellOf := func(x, y int32) (uint32, uint32) { return uint32((y + 8) / 16), uint32((x + 8) / 16) }
		plateau := func(x, y int32) bool {
			r, c := cellOf(x, y)
			tl := baseTile(g, r, c)
			return tl != nil && tl.AtlasKey == "grass_terains" && tl.FrameIndex == 70
		}
		lowGround := func(x, y int32) bool {
			r, c := cellOf(x, y)
			tl := baseTile(g, r, c)
			return tl == nil || !((tl.AtlasKey == "grass_terains" && hillTiles[tl.FrameIndex]) || tl.AtlasKey == "steps")
		}
		onSteps := func(x, y int32) bool {
			for _, yy := range []int32{y, y + 15} {
				for _, xx := range []int32{x + 4, x + 11} {
					if tl := baseTile(g, uint32(yy/16), uint32(xx/16)); tl != nil && tl.AtlasKey == "steps" {
						return true
					}
				}
			}
			return false
		}

		type pos struct{ x, y int32 }
		seen := map[pos]bool{}
		var queue []pos
		maxX, maxY := int32(g.ColCount*16-16), int32(g.RowCount*16-16)
		seed := func(x, y int32) {
			p := pos{x, y}
			if !seen[p] && lowGround(x, y) && s.isTraversableAt(rp, x, y) {
				seen[p] = true
				queue = append(queue, p)
			}
		}
		for i := int32(0); i <= maxX || i <= maxY; i += 2 {
			if i <= maxX {
				seed(i, 0)
				seed(i, maxY)
			}
			if i <= maxY {
				seed(0, i)
				seed(maxX, i)
			}
		}
		var leaks []string
		for len(queue) > 0 && len(leaks) < 5 {
			cur := queue[0]
			queue = queue[1:]
			for _, d := range []pos{{2, 0}, {-2, 0}, {0, 2}, {0, -2}} {
				n := pos{cur.x + d.x, cur.y + d.y}
				if seen[n] || onSteps(n.x, n.y) || !s.isTraversableAt(rp, n.x, n.y) {
					continue
				}
				seen[n] = true
				queue = append(queue, n)
				if plateau(n.x, n.y) {
					r, c := cellOf(n.x, n.y)
					leaks = append(leaks, fmt.Sprintf("plateau (%d,%d) at px (%d,%d) from (%d,%d)", r, c, n.x, n.y, cur.x, cur.y))
				}
			}
		}
		if len(leaks) > 0 {
			t.Errorf("%s: hill reachable without steps: %v", grid, leaks)
		}
	}
}
