package gridgen

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/datlib/tiled"
)

// Palette maps the configured terrain stack onto tiles in a corner wangset.
//
// Level 0 is the base terrain and only needs a full tile. Every higher level
// L is drawn on its own layer using tiles whose corners are either terrain L
// or 0 (transparent), so it needs those "overlay" transitions.
type Palette struct {
	Terrains []string
	colors   []int // wangset color index per level
	// overlay[L][mask] lists tile ids for level L, where mask bit i is set
	// when corner i (NW, NE, SE, SW) is at or above level L.
	overlay []map[uint8][]int
	// weight is each tile's Tiled probability, used to choose between
	// variants that fit the same pattern.
	weight map[int]float64
	// Stale lists wangset tiles that were skipped because they are fully
	// transparent in the tileset image.
	Stale []int
}

const fullMask = uint8(0b1111)

// LoadPalette resolves terrains against a wangset. Tiles in skip (typically
// tiled.BlankTiles) are ignored and reported in Palette.Stale.
func LoadPalette(ws *tiled.WangSet, terrains []string, skip map[int]bool) (*Palette, error) {
	if ws.Type != "corner" && ws.Type != "mixed" {
		return nil, fmt.Errorf("wangset %q is %q; a corner set is required", ws.Name, ws.Type)
	}

	p := &Palette{Terrains: terrains, weight: ws.Probabilities()}
	byCorners := ws.TilesByCorners()
	staleSeen := map[int]bool{}
	for c, ids := range byCorners {
		kept := ids[:0:0]
		for _, id := range ids {
			if skip[id] {
				if !staleSeen[id] {
					staleSeen[id] = true
					p.Stale = append(p.Stale, id)
				}
				continue
			}
			kept = append(kept, id)
		}
		byCorners[c] = kept
	}
	sort.Ints(p.Stale)

	for level, name := range terrains {
		color, ok := ws.ColorIndex(name)
		if !ok {
			return nil, fmt.Errorf("terrain %q is not a color in wangset %q (have %s)", name, ws.Name, colorNames(ws))
		}
		p.colors = append(p.colors, color)

		masks := map[uint8][]int{}
		for mask := uint8(1); mask <= fullMask; mask++ {
			var c tiled.Corners
			for i := 0; i < 4; i++ {
				if mask&(1<<i) != 0 {
					c[i] = color
				}
			}
			if ids := byCorners[c]; len(ids) > 0 {
				masks[mask] = ids
			}
		}

		if len(masks[fullMask]) == 0 {
			return nil, fmt.Errorf("terrain %q has no full tile in wangset %q", name, ws.Name)
		}
		if level > 0 && len(masks) == 1 {
			return nil, fmt.Errorf("terrain %q has no transparent edge tiles in wangset %q, so it cannot be layered over %q",
				name, ws.Name, terrains[level-1])
		}
		p.overlay = append(p.overlay, masks)
	}

	return p, nil
}

func colorNames(ws *tiled.WangSet) string {
	names := make([]string, 0, len(ws.Colors))
	for _, c := range ws.Colors {
		names = append(names, c.Name)
	}
	return strings.Join(names, ", ")
}

func (p *Palette) Levels() int {
	return len(p.Terrains)
}

func (p *Palette) LevelOf(terrain string) (int, bool) {
	for i, t := range p.Terrains {
		if strings.EqualFold(t, terrain) {
			return i, true
		}
	}
	return 0, false
}

// Has reports whether a level has a tile for the given corner mask.
func (p *Palette) Has(level int, mask uint8) bool {
	if mask == 0 {
		return true
	}
	return len(p.overlay[level][mask]) > 0
}

// Tile picks a tile id (0-based, within the tileset) for a level and mask,
// weighted by each variant's Tiled probability like Tiled's terrain brush.
// Zero-probability tiles are only used when every variant is zero.
func (p *Palette) Tile(level int, mask uint8, rng *rand.Rand) (int, bool) {
	ids := p.overlay[level][mask]
	if len(ids) == 0 {
		return 0, false
	}

	total := 0.0
	for _, id := range ids {
		total += p.weightOf(id)
	}
	if total <= 0 {
		return ids[rng.Intn(len(ids))], true
	}

	pick := rng.Float64() * total
	for _, id := range ids {
		pick -= p.weightOf(id)
		if pick < 0 {
			return id, true
		}
	}
	return ids[len(ids)-1], true
}

func (p *Palette) weightOf(id int) float64 {
	if w, ok := p.weight[id]; ok && w > 0 {
		return w
	}
	if _, ok := p.weight[id]; ok {
		return 0
	}
	return 1
}

// MissingMasks describes which partial corner patterns a level lacks; the
// generator avoids producing them.
func (p *Palette) MissingMasks(level int) []string {
	var out []string
	for mask := uint8(1); mask < fullMask; mask++ {
		if !p.Has(level, mask) {
			out = append(out, maskString(mask))
		}
	}
	sort.Strings(out)
	return out
}

func maskString(mask uint8) string {
	names := []string{"NW", "NE", "SE", "SW"}
	var set []string
	for i, n := range names {
		if mask&(1<<i) != 0 {
			set = append(set, n)
		}
	}
	return strings.Join(set, "+")
}
