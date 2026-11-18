package tiled

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WangSet is a Tiled terrain set ("wangset") from a .tsx tileset. Colors are
// 1-indexed in WangTile corner/edge values; 0 means "no terrain".
type WangSet struct {
	Name   string
	Type   string // "corner", "edge" or "mixed"
	Colors []WangColor
	Tiles  []WangTile
}

type WangColor struct {
	Name        string
	Probability float64
}

// WangTile tags one tile with terrain colors. WangID follows Tiled's order:
// top, top-right, right, bottom-right, bottom, bottom-left, left, top-left.
type WangTile struct {
	TileID int
	WangID [8]int
	// Probability is the tile's Tiled "probability" (relative weight among
	// tiles that fit the same pattern). Tiled's default is 1.
	Probability float64
}

// Corners in clockwise order from the top-left: NW, NE, SE, SW.
type Corners [4]int

const (
	CornerNW = 0
	CornerNE = 1
	CornerSE = 2
	CornerSW = 3
)

func (t WangTile) Corners() Corners {
	return Corners{t.WangID[7], t.WangID[1], t.WangID[3], t.WangID[5]}
}

// ColorIndex returns the 1-based color index for a terrain name.
func (ws *WangSet) ColorIndex(name string) (int, bool) {
	for i, c := range ws.Colors {
		if strings.EqualFold(c.Name, name) {
			return i + 1, true
		}
	}
	return 0, false
}

// Probabilities maps tile id to its Tiled probability for every tile in the set.
func (ws *WangSet) Probabilities() map[int]float64 {
	out := make(map[int]float64, len(ws.Tiles))
	for _, t := range ws.Tiles {
		out[t.TileID] = t.Probability
	}
	return out
}

// TilesByCorners indexes the set's tiles by their corner colors. Several tiles
// may share a pattern (variants).
func (ws *WangSet) TilesByCorners() map[Corners][]int {
	out := map[Corners][]int{}
	for _, t := range ws.Tiles {
		c := t.Corners()
		out[c] = append(out[c], t.TileID)
	}
	return out
}

type tsxWangFile struct {
	XMLName  xml.Name          `xml:"tileset"`
	Tiles    []tsxTileWeighted `xml:"tile"`
	WangSets []tsxWangSet      `xml:"wangsets>wangset"`
}

type tsxTileWeighted struct {
	ID          int      `xml:"id,attr"`
	Probability *float64 `xml:"probability,attr"`
}

type tsxWangSet struct {
	Name   string         `xml:"name,attr"`
	Type   string         `xml:"type,attr"`
	Colors []tsxWangColor `xml:"wangcolor"`
	Tiles  []tsxWangTile  `xml:"wangtile"`
}

type tsxWangColor struct {
	Name        string  `xml:"name,attr"`
	Probability float64 `xml:"probability,attr"`
}

type tsxWangTile struct {
	TileID int    `xml:"tileid,attr"`
	WangID string `xml:"wangid,attr"`
}

// LoadWangSets reads every terrain set from a .tsx file.
func LoadWangSets(tsxPath string) ([]WangSet, error) {
	data, err := os.ReadFile(tsxPath)
	if err != nil {
		return nil, err
	}

	var doc tsxWangFile
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", tsxPath, err)
	}

	probability := map[int]float64{}
	for _, t := range doc.Tiles {
		if t.Probability != nil {
			probability[t.ID] = *t.Probability
		}
	}

	sets := make([]WangSet, 0, len(doc.WangSets))
	for _, raw := range doc.WangSets {
		ws := WangSet{Name: raw.Name, Type: raw.Type}
		for _, c := range raw.Colors {
			ws.Colors = append(ws.Colors, WangColor{Name: c.Name, Probability: c.Probability})
		}
		for _, t := range raw.Tiles {
			id, err := parseWangID(t.WangID)
			if err != nil {
				return nil, fmt.Errorf("%s: wangset %q tile %d: %w", tsxPath, raw.Name, t.TileID, err)
			}
			p, ok := probability[t.TileID]
			if !ok {
				p = 1
			}
			ws.Tiles = append(ws.Tiles, WangTile{TileID: t.TileID, WangID: id, Probability: p})
		}
		sets = append(sets, ws)
	}

	return sets, nil
}

// LoadWangSet reads the named terrain set from a .tsx file.
func LoadWangSet(tsxPath string, name string) (*WangSet, error) {
	sets, err := LoadWangSets(tsxPath)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(sets))
	for i := range sets {
		if sets[i].Name == name {
			return &sets[i], nil
		}
		names = append(names, sets[i].Name)
	}

	return nil, fmt.Errorf("%s: no wangset %q (have %v)", tsxPath, name, names)
}

func parseWangID(raw string) ([8]int, error) {
	var out [8]int
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) != 8 {
		return out, fmt.Errorf("wangid %q: want 8 comma-separated values (Tiled 1.5+ format)", raw)
	}
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return out, fmt.Errorf("wangid %q: %w", raw, err)
		}
		out[i] = v
	}
	return out, nil
}
