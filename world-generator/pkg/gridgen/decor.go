package gridgen

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// decorTileset lists the tiles of a .tsx marked with the property decor=true
// (e.g. the generated flora tileset). They are scattered on a "decor" layer
// above terrain; they never collide, and the client animates them.
type decorTileset struct {
	TileCount int
	IDs       []int
}

func loadDecorTileset(path string) (*decorTileset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ts struct {
		TileCount int `xml:"tilecount,attr"`
		Tiles     []struct {
			ID    int `xml:"id,attr"`
			Props []struct {
				Name  string `xml:"name,attr"`
				Value string `xml:"value,attr"`
			} `xml:"properties>property"`
		} `xml:"tile"`
	}
	if err := xml.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	d := &decorTileset{TileCount: ts.TileCount}
	for _, t := range ts.Tiles {
		for _, p := range t.Props {
			if strings.EqualFold(p.Name, "decor") && strings.EqualFold(p.Value, "true") {
				d.IDs = append(d.IDs, t.ID)
			}
		}
	}
	if len(d.IDs) == 0 {
		return nil, fmt.Errorf("%s: no tiles with decor=true", path)
	}
	return d, nil
}

// tsxTileCount reads a tileset's tilecount (gids it occupies).
func tsxTileCount(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var ts struct {
		TileCount int `xml:"tilecount,attr"`
	}
	if err := xml.Unmarshal(data, &ts); err != nil {
		return 0, fmt.Errorf("%s: %w", path, err)
	}
	return ts.TileCount, nil
}

// decorLayer places a random decor tile on each free all-level cell with
// probability density. Returns nil when nothing was placed.
func (m *mapBuilder) decorLayer(level int, density float64, d *decorTileset, firstGID int, used map[cellPos]bool) []uint32 {
	data := make([]uint32, m.cfg.MapRows*m.cfg.MapCols)
	placed := 0
	for _, p := range m.cellsAtLevel(level) {
		if used[p] || m.rng.Float64() >= density {
			continue
		}
		id := d.IDs[m.rng.Intn(len(d.IDs))]
		data[p.row*m.cfg.MapCols+p.col] = uint32(firstGID + id)
		placed++
	}
	if placed == 0 {
		return nil
	}
	return data
}
