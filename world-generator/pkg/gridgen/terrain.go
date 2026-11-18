package gridgen

import (
	"github.com/aquilax/go-perlin"
)

// cornerField holds one terrain level per tile corner across the whole world,
// so neighbouring maps share their edge corners and line up seamlessly.
type cornerField struct {
	rows, cols int // corner counts: tiles + 1
	level      []int
}

func newCornerField(tileRows, tileCols int) *cornerField {
	return &cornerField{
		rows:  tileRows + 1,
		cols:  tileCols + 1,
		level: make([]int, (tileRows+1)*(tileCols+1)),
	}
}

func (f *cornerField) at(r, c int) int {
	return f.level[r*f.cols+c]
}

func (f *cornerField) set(r, c, v int) {
	f.level[r*f.cols+c] = v
}

// cellCorners returns the corner levels of tile (r, c) as NW, NE, SE, SW.
func (f *cornerField) cellCorners(r, c int) [4]int {
	return [4]int{f.at(r, c), f.at(r, c+1), f.at(r+1, c+1), f.at(r+1, c)}
}

var cornerOffsets = [4][2]int{{0, 0}, {0, 1}, {1, 1}, {1, 0}}

func maskAtLeast(corners [4]int, level int) uint8 {
	var m uint8
	for i, v := range corners {
		if v >= level {
			m |= 1 << i
		}
	}
	return m
}

const (
	perlinAlpha      = 2.
	perlinBeta       = 2.
	perlinIterations = int32(3)
)

func fillFromNoise(f *cornerField, cfg *GenerationConfig) {
	noise := perlin.NewPerlin(perlinAlpha, perlinBeta, perlinIterations, cfg.Seed)

	for r := 0; r < f.rows; r++ {
		for c := 0; c < f.cols; c++ {
			v := noise.Noise2D(float64(c)/cfg.NoiseScale, float64(r)/cfg.NoiseScale)
			f.set(r, c, levelForNoise(v, cfg.Thresholds))
		}
	}
}

func levelForNoise(v float64, thresholds []float64) int {
	level := 0
	for _, t := range thresholds {
		if v >= t {
			level++
		}
	}
	return level
}

// fixUnsupported raises corners until every cell's pattern on every layer has
// a tile in the palette (for example, the diagonal checkerboard patterns most
// tilesets omit). Raising only ever increases levels and a full tile always
// exists, so this terminates.
func fixUnsupported(f *cornerField, p *Palette) int {
	raised := 0
	for changed := true; changed; {
		changed = false
		for r := 0; r < f.rows-1; r++ {
			for c := 0; c < f.cols-1; c++ {
				for level := 1; level < p.Levels(); level++ {
					corners := f.cellCorners(r, c)
					mask := maskAtLeast(corners, level)
					if mask == 0 || mask == fullMask || p.Has(level, mask) {
						continue
					}
					for i, off := range cornerOffsets {
						if corners[i] < level {
							f.set(r+off[0], c+off[1], level)
							raised++
						}
					}
					changed = true
				}
			}
		}
	}
	return raised
}
