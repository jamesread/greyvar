// Package faces makes the south faces of raised terrain (hills, walls)
// taller, without redrawing the hand-made tilesets or the maps' terrain.
//
// The source of truth is res/sprites/faces/faces.yml (see Spec).
//
// A "face tile" is a tile whose south edge drops from the terrain to lower
// ground and whose bottom rows show the face (by colour). Each face column
// grows by `extend` pixels: rows Repeat[0]..Repeat[1] of that column repeat,
// then the tile's own bottom rows (the base outline) follow. So every variant
// (over sand, over water, notches, steps) keeps its own colours and shape.
//
// Two overlay tiles per face tile are generated into one tileset:
//
//	cap        drawn over the face tile itself: replaces its bottom outline
//	           with more face (no collision; the face tile has its own)
//	extension  drawn over the cell below: the rest of the face, with
//	           collision (overlay=true adds it to the cell's own collision)
//
// For each listed map, two locked layers ("face caps (generated)" and
// "faces (generated)") are written above the map's tile layers, placing a
// cap on every face tile and an extension below it. Re-run after editing a
// map's terrain (`make sprites`); `make sprites-check` flags stale maps.
package faces

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
	"gopkg.in/yaml.v3"
)

const (
	tileSize  = 16
	CapLayer  = "face caps (generated)"
	FaceLayer = "faces (generated)"
)

// Spec is faces.yml.
type Spec struct {
	// Out is the generated tileset, relative to the spec dir, without
	// extension (writes Out.png and Out.tsx).
	Out string `yaml:"out"`
	// Extend: how many pixels taller every face gets (< 16).
	Extend  int      `yaml:"extend"`
	Columns int      `yaml:"columns,omitempty"`
	Sources []Source `yaml:"sources"`
	// Maps (.tmj, relative to the spec dir) that get the generated layers.
	Maps []string `yaml:"maps"`
}

// Source is a tileset whose face tiles grow.
type Source struct {
	Tileset string `yaml:"tileset"`
	// Wangset/Terrain pick candidate tiles: terrain on a top corner, lower
	// ground on a bottom corner. Or list Tiles explicitly.
	Wangset string `yaml:"wangset,omitempty"`
	Terrain string `yaml:"terrain,omitempty"`
	Tiles   []int  `yaml:"tiles,omitempty"`
	// FaceColors mark face pixels; OutlineColors the face's side outline.
	FaceColors    []string `yaml:"faceColors"`
	OutlineColors []string `yaml:"outlineColors"`
	// Stairs: other opaque columns in the face are walkable stairs; they
	// grow too, but without collision.
	Stairs bool `yaml:"stairs,omitempty"`
	// Repeat is the [first, last] face row that repeats; the last is also
	// the row whose colours classify each column.
	Repeat [2]int `yaml:"repeat"`
}

// Faces is a loaded spec.
type Faces struct {
	Dir  string
	Spec Spec
}

func Load(dir string) (*Faces, error) {
	data, err := os.ReadFile(filepath.Join(dir, "faces.yml"))
	if err != nil {
		return nil, err
	}
	f := &Faces{Dir: dir}
	if err := yaml.Unmarshal(data, &f.Spec); err != nil {
		return nil, fmt.Errorf("faces.yml: %w", err)
	}
	if f.Spec.Columns == 0 {
		f.Spec.Columns = 8
	}
	return f, nil
}

func (f *Faces) path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Clean(filepath.Join(f.Dir, rel))
}

// Validate checks the spec.
func (f *Faces) Validate() []string {
	var errs []string
	s := f.Spec
	if s.Out == "" {
		errs = append(errs, "out is required")
	}
	if s.Extend <= 0 || s.Extend >= tileSize {
		errs = append(errs, fmt.Sprintf("extend must be 1..%d", tileSize-1))
	}
	for i, src := range s.Sources {
		where := fmt.Sprintf("sources[%d]", i)
		if src.Tileset == "" {
			errs = append(errs, where+": tileset is required")
		}
		if len(src.Tiles) == 0 && (src.Wangset == "" || src.Terrain == "") {
			errs = append(errs, where+": give tiles, or wangset and terrain")
		}
		if len(src.FaceColors) == 0 {
			errs = append(errs, where+": faceColors is required")
		}
		for _, c := range append(append([]string{}, src.FaceColors...), src.OutlineColors...) {
			if _, err := rig.ParseHex(c); err != nil {
				errs = append(errs, fmt.Sprintf("%s: colour %q: %v", where, c, err))
			}
		}
		if src.Repeat[0] < 0 || src.Repeat[1] < src.Repeat[0] || src.Repeat[1] >= tileSize-1 {
			errs = append(errs, where+": repeat must be [first, last] within the tile, above its bottom row")
		}
	}
	return errs
}

// column classes
const (
	colNone = iota
	colFace
	colOutline
	colStair
)

// FaceTile is one source tile that grows, with its generated overlays.
type FaceTile struct {
	Tileset string // cleaned path of the source .tsx
	ID      int
	Cap     *image.NRGBA
	Ext     *image.NRGBA
	// ExtCollision is the extension's blocking rects (x, y, w, h).
	ExtCollision [][4]int
}

type tsxInfo struct {
	columns  int
	count    int
	image    *image.NRGBA
	decor    map[int]bool
	wangsets map[string]wangset
	// faceBelows: wall tiles whose face continues into the cell below
	// ({full, short} lower tile ids; see walls.go).
	faceBelows map[int][2]int
}

type wangset struct {
	colors  []string       // index 0 = wang colour 1
	corners map[int][4]int // tile -> TL, TR, BR, BL
}

var (
	reColumns   = regexp.MustCompile(`<tileset [^>]*columns="(\d+)"`)
	reCount     = regexp.MustCompile(`<tileset [^>]*tilecount="(\d+)"`)
	reImage     = regexp.MustCompile(`<image source="([^"]+)"`)
	reWangset   = regexp.MustCompile(`(?s)<wangset name="([^"]+)"[^>]*>(.*?)</wangset>`)
	reWangColor = regexp.MustCompile(`<wangcolor name="([^"]*)"`)
	reWangTile  = regexp.MustCompile(`<wangtile tileid="(\d+)" wangid="([\d,]+)"/>`)
	reTileBlock = regexp.MustCompile(`(?s)<tile id="(\d+)"(?:[^>/]|/[^>])*>(.*?)</tile>`)
	reDecor     = regexp.MustCompile(`<property name="decor"[^>]*value="true"`)
)

func readTsx(path string) (*tsxInfo, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tsx := string(raw)
	info := &tsxInfo{decor: map[int]bool{}, wangsets: map[string]wangset{}, faceBelows: faceBelows(tsx)}
	m := reColumns.FindStringSubmatch(tsx)
	c := reCount.FindStringSubmatch(tsx)
	im := reImage.FindStringSubmatch(tsx)
	if m == nil || c == nil || im == nil {
		return nil, fmt.Errorf("%s: missing columns, tilecount or image", path)
	}
	info.columns, _ = strconv.Atoi(m[1])
	info.count, _ = strconv.Atoi(c[1])
	if info.image, err = rig.ReadPNG(filepath.Join(filepath.Dir(path), filepath.FromSlash(im[1]))); err != nil {
		return nil, err
	}
	for _, b := range reTileBlock.FindAllStringSubmatch(tsx, -1) {
		if reDecor.MatchString(b[2]) {
			id, _ := strconv.Atoi(b[1])
			info.decor[id] = true
		}
	}
	for _, ws := range reWangset.FindAllStringSubmatch(tsx, -1) {
		w := wangset{corners: map[int][4]int{}}
		for _, col := range reWangColor.FindAllStringSubmatch(ws[2], -1) {
			w.colors = append(w.colors, col[1])
		}
		for _, wt := range reWangTile.FindAllStringSubmatch(ws[2], -1) {
			id, _ := strconv.Atoi(wt[1])
			parts := strings.Split(wt[2], ",")
			if len(parts) != 8 {
				continue
			}
			v := make([]int, 8)
			for i, p := range parts {
				v[i], _ = strconv.Atoi(p)
			}
			w.corners[id] = [4]int{v[7], v[1], v[3], v[5]}
		}
		info.wangsets[ws[1]] = w
	}
	return info, nil
}

func (t *tsxInfo) tile(id int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, tileSize, tileSize))
	ox, oy := (id%t.columns)*tileSize, (id/t.columns)*tileSize
	for y := 0; y < tileSize; y++ {
		for x := 0; x < tileSize; x++ {
			out.SetNRGBA(x, y, t.image.NRGBAAt(ox+x, oy+y))
		}
	}
	return out
}

func colorSet(hexes []string) map[rig.RGB]bool {
	set := map[rig.RGB]bool{}
	for _, h := range hexes {
		if c, err := rig.ParseHex(h); err == nil {
			set[c] = true
		}
	}
	return set
}

// candidates are the tiles of src that may have a south face.
func (src Source) candidates(info *tsxInfo) ([]int, error) {
	if len(src.Tiles) > 0 {
		return append([]int(nil), src.Tiles...), nil
	}
	ws, ok := info.wangsets[src.Wangset]
	if !ok {
		return nil, fmt.Errorf("%s: no wangset %q", src.Tileset, src.Wangset)
	}
	terrain := 0
	for i, name := range ws.colors {
		if name == src.Terrain {
			terrain = i + 1
		}
	}
	if terrain == 0 {
		return nil, fmt.Errorf("%s: wangset %q has no colour %q", src.Tileset, src.Wangset, src.Terrain)
	}
	var ids []int
	for id, c := range ws.corners {
		top := c[0] == terrain || c[1] == terrain
		dropsBelow := c[2] != terrain || c[3] != terrain
		if top && dropsBelow {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	return ids, nil
}

func rgbAt(img *image.NRGBA, x, y int) (rig.RGB, bool) {
	c := img.NRGBAAt(x, y)
	return rig.RGB{R: c.R, G: c.G, B: c.B}, c.A > 0
}

// classify each column of a tile by its colour on the probe row.
func (src Source) classify(t *image.NRGBA) []int {
	face, outline := colorSet(src.FaceColors), colorSet(src.OutlineColors)
	probe := src.Repeat[1]
	cls := make([]int, tileSize)
	for x := 0; x < tileSize; x++ {
		c, opaque := rgbAt(t, x, probe)
		switch {
		case !opaque:
		case face[c]:
			cls[x] = colFace
		case outline[c]:
			cls[x] = colOutline
		case src.Stairs:
			cls[x] = colStair
		}
	}
	// An outline only counts beside the face (not e.g. a rim's own line).
	grows := func(x int) bool { return x >= 0 && x < tileSize && (cls[x] == colFace || cls[x] == colStair) }
	for x := range cls {
		if cls[x] == colOutline && !grows(x-1) && !grows(x+1) {
			cls[x] = colNone
		}
	}
	return cls
}

// grow builds the cap and extension for one tile; nil if it has no face.
func (src Source) grow(t *image.NRGBA, extend int) *FaceTile {
	cls := src.classify(t)
	hasFace := false
	for _, c := range cls {
		hasFace = hasFace || c == colFace || c == colStair
	}
	if !hasFace {
		return nil
	}
	first, last := src.Repeat[0], src.Repeat[1]
	period := last - first + 1
	// Virtual row v of the taller face, from the row after the probe down
	// to the new bottom (15 + extend), maps to a row of the original tile.
	srcRow := func(v int) int {
		if v <= last+extend {
			return first + (v-first)%period
		}
		return v - extend
	}
	ft := &FaceTile{
		Cap: image.NewNRGBA(image.Rect(0, 0, tileSize, tileSize)),
		Ext: image.NewNRGBA(image.Rect(0, 0, tileSize, tileSize)),
	}
	for x, c := range cls {
		if c == colNone {
			continue
		}
		for v := last + 1; v < tileSize+extend; v++ {
			px := t.NRGBAAt(x, srcRow(v))
			if v < tileSize {
				ft.Cap.SetNRGBA(x, v, px)
			} else {
				ft.Ext.SetNRGBA(x, v-tileSize, px)
			}
		}
	}
	// Blocking rects: runs of face/outline columns (stairs stay walkable).
	for x := 0; x < tileSize; {
		if cls[x] != colFace && cls[x] != colOutline {
			x++
			continue
		}
		start := x
		for x < tileSize && (cls[x] == colFace || cls[x] == colOutline) {
			x++
		}
		ft.ExtCollision = append(ft.ExtCollision, [4]int{start, 0, x - start, extend})
	}
	return ft
}

// Result is everything Build computed.
type Result struct {
	Tiles   []*FaceTile
	Outputs []rig.Output
	// Placed counts face tiles per map that got a cap and extension.
	Placed map[string]int
}

// Build generates the overlay tileset and the maps' generated layers.
func (f *Faces) Build() (*Result, error) {
	if errs := f.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("faces.yml invalid:\n  %s", strings.Join(errs, "\n  "))
	}
	res := &Result{Placed: map[string]int{}}
	index := map[string]int{} // "tsx#id" -> FaceTiles index
	for _, src := range f.Spec.Sources {
		tsxPath := f.path(src.Tileset)
		info, err := readTsx(tsxPath)
		if err != nil {
			return nil, err
		}
		ids, err := src.candidates(info)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if id < 0 || id >= info.count {
				return nil, fmt.Errorf("%s: tile %d out of range", src.Tileset, id)
			}
			ft := src.grow(info.tile(id), f.Spec.Extend)
			if ft == nil {
				continue
			}
			ft.Tileset, ft.ID = tsxPath, id
			key := fmt.Sprintf("%s#%d", tsxPath, id)
			if _, dup := index[key]; dup {
				continue
			}
			index[key] = len(res.Tiles)
			res.Tiles = append(res.Tiles, ft)
		}
	}
	if len(res.Tiles) == 0 {
		return nil, fmt.Errorf("faces: no face tiles found")
	}

	outBase := f.path(f.Spec.Out)
	cols := f.Spec.Columns
	count := 2 * len(res.Tiles)
	rows := (count + cols - 1) / cols
	sheet := image.NewNRGBA(image.Rect(0, 0, cols*tileSize, rows*tileSize))
	put := func(id int, img *image.NRGBA) {
		ox, oy := (id%cols)*tileSize, (id/cols)*tileSize
		for y := 0; y < tileSize; y++ {
			for x := 0; x < tileSize; x++ {
				sheet.SetNRGBA(ox+x, oy+y, img.NRGBAAt(x, y))
			}
		}
	}
	for i, ft := range res.Tiles {
		put(2*i, ft.Cap)
		put(2*i+1, ft.Ext)
	}
	res.Outputs = append(res.Outputs,
		rig.Output{Path: outBase + ".png", Data: rig.EncodePNG(sheet)},
		rig.Output{Path: outBase + ".tsx", Data: []byte(f.tsx(res.Tiles, outBase, cols, rows))},
	)

	for _, m := range f.Spec.Maps {
		mapPath := f.path(m)
		out, placed, err := f.applyMap(mapPath, outBase+".tsx", count, index)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", mapPath, err)
		}
		res.Placed[mapPath] = placed
		res.Outputs = append(res.Outputs, rig.Output{Path: mapPath, Data: out})
	}
	return res, nil
}

func (f *Faces) tsx(tiles []*FaceTile, outBase string, cols, rows int) string {
	var b strings.Builder
	name := filepath.Base(outBase)
	fmt.Fprintf(&b, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(&b, "<!-- Generated by `spritegen tiles faces` from res/sprites/faces/faces.yml; do not edit.\n")
	fmt.Fprintf(&b, "     Even ids: caps over a face tile. Odd ids: the face's extension into the cell below. -->\n")
	fmt.Fprintf(&b, "<tileset version=\"1.10\" tiledversion=\"1.12.2\" name=\"%s\" tilewidth=\"%d\" tileheight=\"%d\" tilecount=\"%d\" columns=\"%d\">\n",
		name, tileSize, tileSize, 2*len(tiles), cols)
	fmt.Fprintf(&b, " <image source=\"%s.png\" width=\"%d\" height=\"%d\"/>\n", name, cols*tileSize, rows*tileSize)
	for i, ft := range tiles {
		from := fmt.Sprintf("%s#%d", strings.TrimSuffix(filepath.Base(ft.Tileset), ".tsx"), ft.ID)
		for k, part := range []string{"cap", "extension"} {
			fmt.Fprintf(&b, " <tile id=\"%d\">\n  <properties>\n", 2*i+k)
			fmt.Fprintf(&b, "   <property name=\"face\" value=\"%s %s\"/>\n", part, from)
			fmt.Fprintf(&b, "   <property name=\"overlay\" type=\"bool\" value=\"true\"/>\n  </properties>\n")
			if k == 1 && len(ft.ExtCollision) > 0 {
				fmt.Fprintf(&b, "  <objectgroup draworder=\"index\" id=\"2\">\n")
				for j, r := range ft.ExtCollision {
					fmt.Fprintf(&b, "   <object id=\"%d\" x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\"/>\n", j+1, r[0], r[1], r[2], r[3])
				}
				fmt.Fprintf(&b, "  </objectgroup>\n")
			}
			fmt.Fprintf(&b, " </tile>\n")
		}
	}
	fmt.Fprintf(&b, "</tileset>\n")
	return b.String()
}

// ---- maps ----

type tmjTileset struct {
	FirstGID  int    `json:"firstgid"`
	Source    string `json:"source"`
	TileCount int    `json:"tilecount"`
}

type tmjLayer struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Data   []int64 `json:"data"`
}

type tmjMap struct {
	Width    int          `json:"width"`
	Height   int          `json:"height"`
	Tilesets []tmjTileset `json:"tilesets"`
	Layers   []tmjLayer   `json:"layers"`
}

const gidMask = 0x1fffffff

func (f *Faces) applyMap(mapPath, outTsx string, count int, index map[string]int) ([]byte, int, error) {
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		return nil, 0, err
	}
	text := string(raw)
	var m tmjMap
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, 0, err
	}
	dir := filepath.Dir(mapPath)
	hasGenerated := false
	for _, l := range m.Layers {
		hasGenerated = hasGenerated || l.Name == CapLayer || l.Name == FaceLayer
	}
	original := raw

	// Tilesets: resolve each, and find (or allocate) ours.
	type ts struct {
		first int
		path  string
		decor map[int]bool
	}
	var sets []ts
	ourFirst, maxEnd := 0, 1
	for i, t := range m.Tilesets {
		s := ts{first: t.FirstGID}
		n := t.TileCount
		if t.Source != "" {
			s.path = filepath.Clean(filepath.Join(dir, filepath.FromSlash(t.Source)))
			if s.path == outTsx {
				ourFirst = t.FirstGID
				if i != len(m.Tilesets)-1 {
					// It may grow; only the last tileset can safely.
					next := m.Tilesets[i+1].FirstGID
					if t.FirstGID+count > next {
						return nil, 0, fmt.Errorf("%s needs %d gids but the next tileset starts at %d; move it last", filepath.Base(outTsx), count, next)
					}
				}
				continue
			}
			info, err := readTsx(s.path)
			if err != nil {
				return nil, 0, err
			}
			s.decor, n = info.decor, info.count
		}
		sets = append(sets, s)
		if t.FirstGID+n > maxEnd {
			maxEnd = t.FirstGID + n
		}
	}
	sort.Slice(sets, func(i, j int) bool { return sets[i].first < sets[j].first })
	if ourFirst == 0 {
		ourFirst = maxEnd
		rel, err := filepath.Rel(dir, outTsx)
		if err != nil {
			return nil, 0, err
		}
		entry := fmt.Sprintf(", \n        {\n         \"firstgid\":%d,\n         \"source\":\"%s\"\n        }",
			ourFirst, strings.ReplaceAll(filepath.ToSlash(rel), "/", `\/`))
		elems, arrEnd, err := arrayElements(text, `"tilesets":`)
		if err != nil {
			return nil, 0, err
		}
		at := arrEnd
		if len(elems) > 0 {
			at = elems[len(elems)-1][1]
		} else {
			entry = strings.TrimPrefix(entry, ", ")
		}
		text = text[:at] + entry + text[at:]
	}
	lookup := func(gid int64) (string, int, bool) {
		g := int(gid & gidMask)
		if g == 0 {
			return "", 0, false
		}
		for i := len(sets) - 1; i >= 0; i-- {
			if g >= sets[i].first {
				if sets[i].path == "" {
					return "", 0, false
				}
				local := g - sets[i].first
				return sets[i].path, local, sets[i].decor[local]
			}
		}
		return "", 0, false
	}

	// The visible (topmost non-decor) tile of each cell decides.
	W, H := m.Width, m.Height
	caps := make([]int64, W*H)
	exts := make([]int64, W*H)
	placed := 0
	for cell := 0; cell < W*H; cell++ {
		for li := len(m.Layers) - 1; li >= 0; li-- {
			l := m.Layers[li]
			if l.Type != "tilelayer" || isGenerated(l.Name) || cell >= len(l.Data) || l.Data[cell] == 0 {
				continue
			}
			path, local, decor := lookup(l.Data[cell])
			if decor {
				continue
			}
			if i, ok := index[fmt.Sprintf("%s#%d", path, local)]; ok {
				caps[cell] = int64(ourFirst + 2*i)
				if below := cell + W; below < W*H {
					exts[below] = int64(ourFirst + 2*i + 1)
				}
				placed++
			}
			break
		}
	}

	if placed == 0 && !hasGenerated {
		text = string(original) // no cliffs: no hill layers (nor their tileset)
	} else {
		// Write (or replace) the two generated layers above the map's own.
		for _, gen := range []struct {
			name string
			data []int64
		}{{CapLayer, caps}, {FaceLayer, exts}} {
			text, err = upsertLayer(text, gen.name, gen.data, W, H)
			if err != nil {
				return nil, 0, err
			}
		}
	}
	// Tall wall faces: their lower halves, under the wall's own layer.
	text, walls, err := applyWallFaces(text, dir, outTsx)
	if err != nil {
		return nil, 0, err
	}
	if placed == 0 && !hasGenerated && walls == 0 && text == string(original) {
		return original, 0, nil
	}
	placed += walls
	var check tmjMap
	if err := json.Unmarshal([]byte(text), &check); err != nil {
		return nil, 0, fmt.Errorf("edited map is not valid JSON: %w", err)
	}
	return []byte(text), placed, nil
}

var reNextLayer = regexp.MustCompile(`"nextlayerid":\s*(\d+)`)

// upsertLayer replaces the named layer's data, or inserts a new locked tile
// layer after the last tile layer, in Tiled's own formatting.
func upsertLayer(text, name string, data []int64, w, h int) (string, error) {
	elems, _, err := arrayElements(text, `"layers":`)
	if err != nil {
		return "", err
	}
	lastTile := -1
	for i, e := range elems {
		var l tmjLayer
		if err := json.Unmarshal([]byte(text[e[0]:e[1]]), &l); err != nil {
			return "", err
		}
		if l.Name == name {
			// e[0] is the "{", after the indentation layerText also writes.
			return text[:e[0]] + strings.TrimLeft(layerText(name, l.ID, data, w, h), " ") + text[e[1]:], nil
		}
		if l.Type == "tilelayer" {
			lastTile = i
		}
	}
	if lastTile < 0 {
		return "", fmt.Errorf("no tile layers")
	}
	nm := reNextLayer.FindStringSubmatchIndex(text)
	if nm == nil {
		return "", fmt.Errorf("missing nextlayerid")
	}
	id, _ := strconv.Atoi(text[nm[2]:nm[3]])
	text = text[:nm[2]] + strconv.Itoa(id+1) + text[nm[3]:]
	elems, _, _ = arrayElements(text, `"layers":`)
	at := elems[lastTile][1]
	return text[:at] + ", \n" + layerText(name, id, data, w, h) + text[at:], nil
}

func layerText(name string, id int, data []int64, w, h int) string {
	var rows []string
	for r := 0; r < h; r++ {
		var vals []string
		for c := 0; c < w; c++ {
			vals = append(vals, strconv.FormatInt(data[r*w+c], 10))
		}
		rows = append(rows, strings.Join(vals, ", "))
	}
	return fmt.Sprintf(`        {
         "data":[%s],
         "height":%d,
         "id":%d,
         "locked":true,
         "name":"%s",
         "opacity":1,
         "type":"tilelayer",
         "visible":true,
         "width":%d,
         "x":0,
         "y":0
        }`, strings.Join(rows, ",\n            "), h, id, name, w)
}

// arrayElements finds the top-level JSON array after key and returns each
// element's [start, end) offsets and the offset of the closing bracket.
func arrayElements(text, key string) ([][2]int, int, error) {
	k := strings.Index(text, key)
	if k < 0 {
		return nil, 0, fmt.Errorf("missing %s", key)
	}
	open := strings.Index(text[k:], "[")
	if open < 0 {
		return nil, 0, fmt.Errorf("%s is not an array", key)
	}
	i := k + open + 1
	var elems [][2]int
	depth, start, inStr := 0, -1, false
	for ; i < len(text); i++ {
		ch := text[i]
		if inStr {
			if ch == '\\' {
				i++
			} else if ch == '"' {
				inStr = false
			}
			continue
		}
		switch ch {
		case '"':
			inStr = true
			if depth == 0 && start < 0 {
				start = i
			}
		case '{', '[':
			if depth == 0 {
				start = i
			}
			depth++
		case '}', ']':
			if depth == 0 { // the array's own ]
				return elems, i, nil
			}
			depth--
			if depth == 0 {
				elems = append(elems, [2]int{start, i + 1})
				start = -1
			}
		}
	}
	return nil, 0, fmt.Errorf("%s: unterminated array", key)
}
