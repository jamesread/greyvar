package tiled

import (
	"encoding/xml"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
)

// BlankTiles returns the ids of tiles in a single-image .tsx tileset whose
// pixels are all fully transparent. Such tiles usually mean the image was
// edited after terrain/wang data was assigned, leaving stale tags behind.
func BlankTiles(tsxPath string) (map[int]bool, error) {
	data, err := os.ReadFile(tsxPath)
	if err != nil {
		return nil, err
	}

	var ts tsxTileset
	if err := xml.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("parse %s: %w", tsxPath, err)
	}
	if ts.Image.Source == "" {
		return nil, fmt.Errorf("%s: not a single-image tileset", tsxPath)
	}
	if ts.TileWidth <= 0 || ts.TileHeight <= 0 || ts.Columns <= 0 {
		return nil, fmt.Errorf("%s: missing tile size or columns", tsxPath)
	}

	imgPath := filepath.Join(filepath.Dir(tsxPath), filepath.FromSlash(ts.Image.Source))
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", imgPath, err)
	}

	count := ts.TileCount
	if count <= 0 {
		b := img.Bounds()
		count = ts.Columns * (b.Dy() / ts.TileHeight)
	}

	blank := map[int]bool{}
	for id := 0; id < count; id++ {
		x0 := (id % ts.Columns) * ts.TileWidth
		y0 := (id / ts.Columns) * ts.TileHeight
		if tileIsBlank(img, x0, y0, ts.TileWidth, ts.TileHeight) {
			blank[id] = true
		}
	}
	return blank, nil
}

func tileIsBlank(img image.Image, x0, y0, w, h int) bool {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a != 0 {
				return false
			}
		}
	}
	return true
}
