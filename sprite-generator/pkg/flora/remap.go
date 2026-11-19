package flora

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// RemapOptions replaces an embedded single-image tile (matched by image file
// name) with a tile from an external .tsx, e.g. legacy water.png cells with
// the animated water tileset.
type RemapOptions struct {
	FromImage string // basename of the embedded tile's image, e.g. water.png
	ToTsx     string // external tileset path
	ToTile    int    // local tile id in ToTsx
}

// RemapEmbeddedTile edits one map in place (same text-preserving, verified
// approach as Migrate). Maps that do not place the tile are left untouched.
func RemapEmbeddedTile(mapPath string, o RemapOptions, write bool) (MigrateResult, error) {
	res := MigrateResult{Path: mapPath}
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		return res, err
	}
	var doc struct {
		Layers []struct {
			Name string   `json:"name"`
			Type string   `json:"type"`
			Data []uint32 `json:"data"`
		} `json:"layers"`
		Tilesets []struct {
			FirstGID uint32 `json:"firstgid"`
			Source   string `json:"source"`
			Tiles    []struct {
				ID    uint32 `json:"id"`
				Image string `json:"image"`
			} `json:"tiles"`
		} `json:"tilesets"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return res, fmt.Errorf("%s: %w", mapPath, err)
	}
	mapDir, err := filepath.Abs(filepath.Dir(mapPath))
	if err != nil {
		return res, err
	}
	toBase := filepath.Base(o.ToTsx)

	from := map[uint32]bool{}
	next := uint32(1)
	for _, ts := range doc.Tilesets {
		if ts.Source != "" {
			if filepath.Base(ts.Source) == toBase {
				res.Skipped = "already references " + toBase
				return res, nil
			}
			count, err := tsxTileCount(filepath.Join(mapDir, filepath.FromSlash(ts.Source)))
			if err != nil {
				return res, err
			}
			if end := ts.FirstGID + uint32(count); end > next {
				next = end
			}
			continue
		}
		for _, t := range ts.Tiles {
			if filepath.Base(t.Image) == o.FromImage {
				from[ts.FirstGID+t.ID] = true
			}
		}
	}
	if e := embeddedEnd(raw); e > next {
		next = e
	}
	newGID := next + uint32(o.ToTile)

	text := string(raw)
	spans, err := arrayElementSpans(text, "layers")
	if err != nil {
		return res, err
	}
	expected := map[int][]uint32{}
	for i := len(doc.Layers) - 1; i >= 0; i-- {
		l := doc.Layers[i]
		if l.Type != "tilelayer" {
			continue
		}
		data := append([]uint32(nil), l.Data...)
		changed := 0
		for k, g := range data {
			if from[g&^gidFlags] {
				data[k] = newGID | (g & gidFlags)
				changed++
			}
		}
		if changed == 0 {
			continue
		}
		layerText, err := replaceData(text[spans[i][0]:spans[i][1]], data)
		if err != nil {
			return res, fmt.Errorf("%s layer %q: %w", mapPath, l.Name, err)
		}
		text = text[:spans[i][0]] + layerText + text[spans[i][1]:]
		expected[i] = data
		res.Moved += changed
		res.Layers = append([]string{l.Name}, res.Layers...)
	}
	if res.Moved == 0 {
		res.Skipped = "does not place " + o.FromImage
		return res, nil
	}

	tsSpans, err := arrayElementSpans(text, "tilesets")
	if err != nil || len(tsSpans) == 0 {
		return res, fmt.Errorf("%s: cannot find tilesets: %v", mapPath, err)
	}
	rel, err := filepath.Rel(mapDir, o.ToTsx)
	if err != nil {
		return res, err
	}
	// Style the new entry like the last entry's opening line.
	last := tsSpans[len(tsSpans)-1]
	entry := fmt.Sprintf(`{"firstgid":%d, "source":%s}`, next, jsonString(filepath.ToSlash(rel)))
	if template := firstExternalEntry(text, tsSpans); template != "" {
		entry = replaceFirst(template, `"firstgid"\s*:\s*\d+`, fmt.Sprintf(`"firstgid":%d`, next))
		entry = replaceFirst(entry, `"source"\s*:\s*"(?:[^"\\]|\\.)*"`, `"source":`+jsonString(filepath.ToSlash(rel)))
	} else {
		indent := text[strings.LastIndex(text[:last[0]], "\n")+1 : last[0]]
		entry = "{\n" + indent + " \"firstgid\":" + fmt.Sprint(next) + ",\n" + indent + " \"source\":" + jsonString(filepath.ToSlash(rel)) + "\n" + indent + "}"
	}
	text = text[:last[1]] + separator(text, tsSpans, len(tsSpans)-1) + entry + text[last[1]:]

	// Verify.
	var check struct {
		Layers []struct {
			Data []uint32 `json:"data"`
		} `json:"layers"`
		Tilesets []struct {
			FirstGID uint32 `json:"firstgid"`
			Source   string `json:"source"`
		} `json:"tilesets"`
	}
	if err := json.Unmarshal([]byte(text), &check); err != nil {
		return res, fmt.Errorf("%s: edit produced invalid JSON: %w", mapPath, err)
	}
	for i, want := range expected {
		if !reflect.DeepEqual(check.Layers[i].Data, want) {
			return res, fmt.Errorf("%s: edit check failed for layer %d", mapPath, i)
		}
	}
	lt := check.Tilesets[len(check.Tilesets)-1]
	if lt.FirstGID != next || filepath.Base(lt.Source) != toBase {
		return res, fmt.Errorf("%s: edit check failed (tileset)", mapPath)
	}
	if write {
		if err := os.WriteFile(mapPath, []byte(text), 0o644); err != nil {
			return res, err
		}
	}
	return res, nil
}

func firstExternalEntry(text string, spans [][2]int) string {
	for _, sp := range spans {
		if t := text[sp[0]:sp[1]]; strings.Contains(t, `"source"`) {
			return t
		}
	}
	return ""
}
