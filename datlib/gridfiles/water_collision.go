package gridfiles

import "github.com/jamesread/greyvar/datlib/tiled"

// Water blocks wherever it shows. A cell's terrain tile gets collision for
//   - its own pixels drawn in water colour (e.g. a lake edge), and
//   - its see-through pixels where the tiles beneath show water there
//     (e.g. a sand or grass edge painted over a water layer).
// The same see-through edge over grass stays walkable. Rects are snapped to
// 4px blocks (a block blocks if at least half of it is water).

const waterBlock = 4

func applyWaterCollision(g *Grid) {
	var terrain []int
	for i, l := range g.Layers {
		if l.Kind == LayerKindTerrain {
			terrain = append(terrain, i)
		}
	}
	for row := uint32(0); row < g.RowCount; row++ {
		for col := uint32(0); col < g.ColCount; col++ {
			k, top := baseTerrainTile(g, terrain, len(terrain)-1, row, col)
			if top == nil || top.Water {
				continue
			}
			var mask tiled.PixelMask
			if top.WaterPixels != nil {
				mask = *top.WaterPixels
			}
			if top.SeeThrough != nil {
				shown := exposedWater(g, terrain, k-1, row, col, *top.SeeThrough)
				for y := range mask {
					mask[y] |= shown[y]
				}
			}
			top.Collision = append(top.Collision, maskRects(mask)...)
		}
	}
}

// baseTerrainTile is the topmost tile at or below terrain[k] that is not
// decor or an overlay, and its index in terrain.
func baseTerrainTile(g *Grid, terrain []int, k int, row, col uint32) (int, *Tile) {
	for ; k >= 0; k-- {
		if t := g.Layers[terrain[k]].TileAt(row, col); t != nil && !t.Decor && !t.Overlay {
			return k, t
		}
	}
	return -1, nil
}

// exposedWater is the part of `through` (pixels seen from above) that shows
// water in the tiles at or below terrain[k].
func exposedWater(g *Grid, terrain []int, k int, row, col uint32, through tiled.PixelMask) tiled.PixelMask {
	var out tiled.PixelMask
	k, t := baseTerrainTile(g, terrain, k, row, col)
	if t == nil || through.Empty() {
		return out
	}
	if t.Water {
		return through
	}
	for y := range out {
		if t.WaterPixels != nil {
			out[y] = through[y] & t.WaterPixels[y]
		}
	}
	if t.SeeThrough != nil {
		var deeper tiled.PixelMask
		for y := range deeper {
			deeper[y] = through[y] & t.SeeThrough[y]
		}
		below := exposedWater(g, terrain, k-1, row, col, deeper)
		for y := range out {
			out[y] |= below[y]
		}
	}
	return out
}

// maskRects turns a pixel mask into few rects on a 4px grid.
func maskRects(mask tiled.PixelMask) []CollisionRect {
	const n = 16 / waterBlock
	var blocks [n][n]bool
	any := false
	for by := 0; by < n; by++ {
		for bx := 0; bx < n; bx++ {
			count := 0
			for y := by * waterBlock; y < (by+1)*waterBlock; y++ {
				for x := bx * waterBlock; x < (bx+1)*waterBlock; x++ {
					if mask[y]&(1<<x) != 0 {
						count++
					}
				}
			}
			blocks[by][bx] = count*2 >= waterBlock*waterBlock
			any = any || blocks[by][bx]
		}
	}
	if !any {
		return nil
	}
	// Horizontal runs per block row, merged downwards when identical.
	type run struct{ x0, x1, y0, y1 int }
	var open, done []run
	for by := 0; by < n; by++ {
		var rowRuns []run
		for bx := 0; bx < n; {
			if !blocks[by][bx] {
				bx++
				continue
			}
			s := bx
			for bx < n && blocks[by][bx] {
				bx++
			}
			rowRuns = append(rowRuns, run{s, bx, by, by + 1})
		}
		var next []run
		for _, r := range rowRuns {
			merged := false
			for i, o := range open {
				if o.x0 == r.x0 && o.x1 == r.x1 && o.y1 == by {
					open[i].y1 = by + 1
					next = append(next, open[i])
					open[i].y1 = -1 // consumed
					merged = true
					break
				}
			}
			if !merged {
				next = append(next, r)
			}
		}
		for _, o := range open {
			if o.y1 != -1 {
				done = append(done, o)
			}
		}
		open = next
	}
	done = append(done, open...)
	rects := make([]CollisionRect, 0, len(done))
	for _, r := range done {
		rects = append(rects, CollisionRect{
			X: float64(r.x0 * waterBlock), Y: float64(r.y0 * waterBlock),
			W: float64((r.x1 - r.x0) * waterBlock), H: float64((r.y1 - r.y0) * waterBlock),
		})
	}
	return rects
}
