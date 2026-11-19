package flora

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Tiled GID flip flags (top three bits); preserved when a tile moves.
const gidFlags = 0xE0000000

// MigrateResult describes what happened to one map.
type MigrateResult struct {
	Path    string
	Moved   int      // flower tiles moved to decor layers
	Layers  []string // decor layers added
	Skipped string   // non-empty: why the map was left alone
}

type mapDoc struct {
	NextLayerID int `json:"nextlayerid"`
	Layers      []struct {
		ID     int      `json:"id"`
		Name   string   `json:"name"`
		Type   string   `json:"type"`
		Width  int      `json:"width"`
		Height int      `json:"height"`
		Data   []uint32 `json:"data"`
		Props  any      `json:"properties"`
	} `json:"layers"`
	Tilesets []struct {
		FirstGID uint32 `json:"firstgid"`
		Source   string `json:"source"`
	} `json:"tilesets"`
}

// Migrate moves baked flower tiles in one Tiled map onto a decor layer of
// animated flora tiles. It edits the JSON text in place (Tiled's formatting
// survives), then re-parses the result and checks every layer before writing.
// Maps that already reference the flora tileset are skipped, so it is safe to
// re-run.
func (f *Flora) Migrate(mapPath string, write bool) (MigrateResult, error) {
	res := MigrateResult{Path: mapPath}
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		return res, err
	}
	var doc mapDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return res, fmt.Errorf("%s: %w", mapPath, err)
	}
	mapDir := filepath.Dir(mapPath)
	floraTsx := f.path(f.Spec.Outputs.Tsx)
	floraBase := filepath.Base(floraTsx)

	// local id -> flora index, per source tileset basename
	sources := map[string]map[int]int{}
	for i, e := range f.Spec.Flora {
		for _, src := range e.Sources {
			ts, idStr, ok := strings.Cut(src, "#")
			id, err := strconv.Atoi(idStr)
			if !ok || err != nil {
				return res, fmt.Errorf("flora %s: bad source %q (want tileset.tsx#id)", e.Name, src)
			}
			if sources[ts] == nil {
				sources[ts] = map[int]int{}
			}
			sources[ts][id] = i
		}
	}

	nextGID := uint32(1)
	type srcSet struct {
		first       uint32
		ids         map[int]int
		replacement uint32
	}
	var sets []srcSet
	for _, ts := range doc.Tilesets {
		if ts.Source == "" {
			continue // embedded: no baked flowers here
		}
		base := filepath.Base(ts.Source)
		if base == floraBase {
			res.Skipped = "already references " + floraBase
			return res, nil
		}
		count, err := tsxTileCount(filepath.Join(mapDir, filepath.FromSlash(ts.Source)))
		if err != nil {
			return res, err
		}
		if end := ts.FirstGID + uint32(count); end > nextGID {
			nextGID = end
		}
		if ids, ok := sources[base]; ok {
			rep, ok := f.Spec.Migrate.Replacement[base]
			if !ok {
				return res, fmt.Errorf("flora.yml migrate.replacement has no entry for %s", base)
			}
			sets = append(sets, srcSet{first: ts.FirstGID, ids: ids, replacement: ts.FirstGID + uint32(rep)})
		}
	}
	// Embedded tilesets also occupy gids; never overlap them.
	if gids := embeddedEnd(raw); gids > nextGID {
		nextGID = gids
	}
	floraFirst := nextGID
	frames := uint32(f.FrameCount())

	type change struct {
		layer         int
		data, decor   []uint32
	}
	var changes []change
	for li, l := range doc.Layers {
		if l.Type != "tilelayer" {
			continue
		}
		data := append([]uint32(nil), l.Data...)
		decor := make([]uint32, len(l.Data))
		moved := 0
		for i, g := range l.Data {
			id, flags := g&^gidFlags, g&gidFlags
			for _, s := range sets {
				if id < s.first {
					continue
				}
				if fi, ok := s.ids[int(id-s.first)]; ok {
					data[i] = s.replacement | flags
					decor[i] = (floraFirst + uint32(fi)*frames) | flags
					moved++
					break
				}
			}
		}
		if moved > 0 {
			if l.Props != nil {
				return res, fmt.Errorf("%s: layer %q has properties; migrate it by hand", mapPath, l.Name)
			}
			changes = append(changes, change{li, data, decor})
			res.Moved += moved
		}
	}
	if len(changes) == 0 {
		res.Skipped = "no baked flowers"
		return res, nil
	}

	text := string(raw)
	layerSpans, err := arrayElementSpans(text, "layers")
	if err != nil {
		return res, fmt.Errorf("%s: %w", mapPath, err)
	}
	nextLayerID := doc.NextLayerID
	expected := map[string][]uint32{}
	// Edit from the last layer backwards so earlier spans stay valid.
	for c := len(changes) - 1; c >= 0; c-- {
		ch := changes[c]
		span := layerSpans[ch.layer]
		layerText := text[span[0]:span[1]]
		newLayer, err := replaceData(layerText, ch.data)
		if err != nil {
			return res, fmt.Errorf("%s layer %q: %w", mapPath, doc.Layers[ch.layer].Name, err)
		}
		name := "decor"
		if len(changes) > 1 {
			name = "decor: " + doc.Layers[ch.layer].Name
		}
		decorLayer, err := replaceData(layerText, ch.decor)
		if err != nil {
			return res, err
		}
		decorLayer = replaceFirst(decorLayer, `"name"\s*:\s*"(?:[^"\\]|\\.)*"`, `"name":`+jsonString(name))
		decorLayer = replaceFirst(decorLayer, `"id"\s*:\s*\d+`, fmt.Sprintf(`"id":%d`, nextLayerID+c))
		sep := separator(text, layerSpans, ch.layer)
		text = text[:span[0]] + newLayer + sep + decorLayer + text[span[1]:]
		expected[doc.Layers[ch.layer].Name] = ch.data
		expected[name] = ch.decor
		res.Layers = append([]string{name}, res.Layers...)
	}
	text = replaceFirst(text, `"nextlayerid"\s*:\s*\d+`, fmt.Sprintf(`"nextlayerid":%d`, nextLayerID+len(changes)))

	// Add the flora tileset, styled like an existing external entry.
	tsSpans, err := arrayElementSpans(text, "tilesets")
	if err != nil || len(tsSpans) == 0 {
		return res, fmt.Errorf("%s: cannot find tilesets: %v", mapPath, err)
	}
	var template string
	for _, sp := range tsSpans {
		if strings.Contains(text[sp[0]:sp[1]], `"source"`) {
			template = text[sp[0]:sp[1]]
		}
	}
	rel, err := filepath.Rel(mapDir, floraTsx)
	if err != nil {
		return res, err
	}
	entry := replaceFirst(template, `"firstgid"\s*:\s*\d+`, fmt.Sprintf(`"firstgid":%d`, floraFirst))
	entry = replaceFirst(entry, `"source"\s*:\s*"(?:[^"\\]|\\.)*"`, `"source":`+jsonString(filepath.ToSlash(rel)))
	last := tsSpans[len(tsSpans)-1]
	text = text[:last[1]] + separator(text, tsSpans, len(tsSpans)-1) + entry + text[last[1]:]

	// Verify the edited document says exactly what we meant.
	var check mapDoc
	if err := json.Unmarshal([]byte(text), &check); err != nil {
		return res, fmt.Errorf("%s: edit produced invalid JSON: %w", mapPath, err)
	}
	if len(check.Layers) != len(doc.Layers)+len(changes) || check.NextLayerID != nextLayerID+len(changes) {
		return res, fmt.Errorf("%s: edit check failed (layers/nextlayerid)", mapPath)
	}
	for _, l := range check.Layers {
		if want, ok := expected[l.Name]; ok && !reflect.DeepEqual(l.Data, want) {
			return res, fmt.Errorf("%s: edit check failed for layer %q", mapPath, l.Name)
		}
	}
	lastTs := check.Tilesets[len(check.Tilesets)-1]
	if lastTs.FirstGID != floraFirst || filepath.Base(lastTs.Source) != floraBase {
		return res, fmt.Errorf("%s: edit check failed (tileset)", mapPath)
	}
	if write {
		if err := os.WriteFile(mapPath, []byte(text), 0o644); err != nil {
			return res, err
		}
	}
	return res, nil
}

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

// embeddedEnd is the first gid after any embedded tileset.
func embeddedEnd(raw []byte) uint32 {
	var doc struct {
		Tilesets []struct {
			FirstGID  uint32 `json:"firstgid"`
			Source    string `json:"source"`
			TileCount uint32 `json:"tilecount"`
			Tiles     []struct {
				ID uint32 `json:"id"`
			} `json:"tiles"`
		} `json:"tilesets"`
	}
	_ = json.Unmarshal(raw, &doc)
	var end uint32
	for _, ts := range doc.Tilesets {
		if ts.Source != "" {
			continue
		}
		n := ts.TileCount
		for _, t := range ts.Tiles {
			if t.ID+1 > n {
				n = t.ID + 1
			}
		}
		if ts.FirstGID+n > end {
			end = ts.FirstGID + n
		}
	}
	return end
}

// arrayElementSpans finds the top-level array value of key and returns the
// [start, end) byte span of each element object.
func arrayElementSpans(text, key string) ([][2]int, error) {
	loc := regexp.MustCompile(`"` + regexp.QuoteMeta(key) + `"\s*:\s*\[`).FindStringIndex(text)
	if loc == nil {
		return nil, fmt.Errorf("no %q array", key)
	}
	var spans [][2]int
	depth, start, inStr := 0, -1, false
	for i := loc[1]; i < len(text); i++ {
		c := text[i]
		if inStr {
			if c == '\\' {
				i++
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{', '[':
			if depth == 0 && c == '{' {
				start = i
			}
			depth++
		case '}', ']':
			if depth == 0 && c == ']' {
				return spans, nil
			}
			depth--
			if depth == 0 && c == '}' {
				spans = append(spans, [2]int{start, i + 1})
			}
		}
	}
	return nil, fmt.Errorf("unterminated %q array", key)
}

// separator is the text between array elements, copied from the document so
// inserted elements match its formatting.
func separator(text string, spans [][2]int, i int) string {
	switch {
	case i+1 < len(spans):
		return text[spans[i][1]:spans[i+1][0]]
	case i > 0:
		return text[spans[i-1][1]:spans[i][0]]
	default:
		// Only element: comma plus the element's own indentation.
		lineStart := strings.LastIndex(text[:spans[i][0]], "\n")
		return ",\n" + text[lineStart+1:spans[i][0]]
	}
}

var numberRe = regexp.MustCompile(`\d+`)

// replaceData swaps the numbers inside "data":[...] keeping all whitespace.
func replaceData(layerText string, data []uint32) (string, error) {
	loc := regexp.MustCompile(`"data"\s*:\s*\[`).FindStringIndex(layerText)
	if loc == nil {
		return "", fmt.Errorf("no data array (compressed or base64 layers are not supported)")
	}
	end := strings.IndexByte(layerText[loc[1]:], ']')
	if end < 0 {
		return "", fmt.Errorf("unterminated data array")
	}
	body := layerText[loc[1] : loc[1]+end]
	n := 0
	out := numberRe.ReplaceAllStringFunc(body, func(string) string {
		v := ""
		if n < len(data) {
			v = strconv.FormatUint(uint64(data[n]), 10)
		}
		n++
		return v
	})
	if n != len(data) {
		return "", fmt.Errorf("data has %d values, expected %d", n, len(data))
	}
	return layerText[:loc[1]] + out + layerText[loc[1]+end:], nil
}

func replaceFirst(text, pattern, repl string) string {
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(text)
	if loc == nil {
		return text
	}
	return text[:loc[0]] + repl + text[loc[1]:]
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
