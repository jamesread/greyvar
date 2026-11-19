package tiled

import (
	"image/png"
	"math/bits"
	"os"
	"sync"
	"time"
)

// PixelMask marks pixels of a 16x16 tile: bit x of row y.
type PixelMask [16]uint16

func (m *PixelMask) Empty() bool {
	if m == nil {
		return true
	}
	for _, r := range m {
		if r != 0 {
			return false
		}
	}
	return true
}

// Count is the number of marked pixels.
func (m *PixelMask) Count() int {
	if m == nil {
		return 0
	}
	n := 0
	for _, r := range m {
		n += bits.OnesCount16(r)
	}
	return n
}

// tileMasks are a tile's see-through (alpha 0) pixels and the pixels drawn
// in the tileset's water colour (the colours of its water=true tiles).
type tileMasks struct {
	seeThrough *PixelMask
	water      *PixelMask
}

type maskCacheEntry struct {
	mod   time.Time
	size  int64
	water string // the water colours the masks were computed for
	masks map[int]tileMasks
}

var (
	maskCacheMu sync.Mutex
	maskCache   = map[string]maskCacheEntry{}
)

// computeTileMasks reads a 16x16 atlas and returns the masks of every tile
// that has any see-through or water-coloured pixels. waterTiles are the ids
// of the tileset's water=true tiles; their colours define "water".
func computeTileMasks(imagePath string, columns int, waterTiles []int) map[int]tileMasks {
	st, err := os.Stat(imagePath)
	if err != nil {
		return nil
	}
	key := ""
	for _, id := range waterTiles {
		key += string(rune(id)) + ","
	}
	maskCacheMu.Lock()
	if e, ok := maskCache[imagePath]; ok && e.mod.Equal(st.ModTime()) && e.size == st.Size() && e.water == key {
		maskCacheMu.Unlock()
		return e.masks
	}
	maskCacheMu.Unlock()

	f, err := os.Open(imagePath)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil
	}
	b := img.Bounds()
	rows := b.Dy() / 16
	if columns <= 0 || columns > b.Dx()/16 {
		columns = b.Dx() / 16
	}
	type rgb struct{ r, g, b uint32 }
	at := func(id, x, y int) (rgb, bool) {
		px := b.Min.X + (id%columns)*16 + x
		py := b.Min.Y + (id/columns)*16 + y
		r, g, bl, a := img.At(px, py).RGBA()
		return rgb{r >> 8, g >> 8, bl >> 8}, a == 0
	}
	waterColors := map[rgb]bool{}
	for _, id := range waterTiles {
		if id/columns >= rows {
			continue
		}
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if c, clear := at(id, x, y); !clear {
					waterColors[c] = true
				}
			}
		}
	}
	out := map[int]tileMasks{}
	for id := 0; id < rows*columns; id++ {
		var see, water PixelMask
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				c, clear := at(id, x, y)
				if clear {
					see[y] |= 1 << x
				} else if waterColors[c] {
					water[y] |= 1 << x
				}
			}
		}
		var tm tileMasks
		if !see.Empty() {
			s := see
			tm.seeThrough = &s
		}
		if !water.Empty() {
			w := water
			tm.water = &w
		}
		if tm.seeThrough != nil || tm.water != nil {
			out[id] = tm
		}
	}
	maskCacheMu.Lock()
	maskCache[imagePath] = maskCacheEntry{mod: st.ModTime(), size: st.Size(), water: key, masks: out}
	maskCacheMu.Unlock()
	return out
}
