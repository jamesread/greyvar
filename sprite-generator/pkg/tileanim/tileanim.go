// Package tileanim animates the shoreline tiles of a hand-made tileset from
// frames made in PixelLab, without touching the hand-drawn art.
//
// The source of truth (res/sprites/<name>/):
//
//	tileanim.yml         config (see Spec)
//	frames/*.png         PixelLab animation of a small map ("composition")
//	                     built from the tileset's own tiles (layout in the yml)
//
// For every tile in the layout that has a waterline, each frame is cut from
// the animation and blended with the original tile: pixels within `band` of
// the tile's own waterline take the animated (palette-mapped) pixel, the rest
// keep the original art, so the tile still meets plain neighbours seamlessly.
// Frames are appended below the source rows of the tileset image and each such
// tile gets an <animation> plus synced=true in the .tsx.
//
// The source rows are read from the top of the image every run, so
// re-exporting the art (which drops the generated rows) and running
// `make sprites` restores the animation; `sprites-check` flags it meanwhile.
package tileanim

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
	"gopkg.in/yaml.v3"
)

// Spec is tileanim.yml.
type Spec struct {
	Tileset    string `yaml:"tileset"`    // .tsx, relative to the spec dir
	SourceRows int    `yaml:"sourceRows"` // hand-drawn rows; generated frames go below
	Ms         int    `yaml:"ms"`         // duration of each frame
	// Synced animations start on the same frame in every cell, so the shore
	// moves as one (water). Default true.
	Synced *bool `yaml:"synced,omitempty"`

	// Sources are PixelLab animations of small compositions of the tileset's
	// own tiles. A tile animates from the first source that has it.
	Sources []Source `yaml:"sources"`
	// WaterColor marks water pixels in the original tiles (finds waterlines).
	WaterColor string `yaml:"waterColor"`
	// Band: pixels within this distance of a tile's waterline animate.
	Band int `yaml:"band"`
	// ColorMap maps animation colours onto the tileset palette: every
	// animated pixel becomes the target of its nearest `from` colour.
	ColorMap []ColorPair `yaml:"colorMap"`
}

// Source is one animated composition.
type Source struct {
	// Frames are the animation images (all the same size as the layout).
	Frames []string `yaml:"frames"`
	// Layout is the tile ids the frames were made from, row by row.
	Layout [][]int `yaml:"layout"`
	// Aliases animate other tiles from a layout tile's cell (same corner
	// shape, e.g. a transparent sand edge from the opaque water/sand one).
	Aliases map[int]int `yaml:"aliases,omitempty"`
}

// ColorPair is one colour mapping.
type ColorPair struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// Anim is a loaded spec.
type Anim struct {
	Dir  string
	Spec Spec
}

// Load reads tileanim.yml.
func Load(dir string) (*Anim, error) {
	data, err := os.ReadFile(filepath.Join(dir, "tileanim.yml"))
	if err != nil {
		return nil, err
	}
	a := &Anim{Dir: dir}
	if err := yaml.Unmarshal(data, &a.Spec); err != nil {
		return nil, fmt.Errorf("tileanim.yml: %w", err)
	}
	return a, nil
}

func (a *Anim) synced() bool { return a.Spec.Synced == nil || *a.Spec.Synced }

func (a *Anim) path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Clean(filepath.Join(a.Dir, rel))
}

type tsxHeader struct {
	tileW, tileH, columns int
	imageSource           string
}

var (
	reTileW   = regexp.MustCompile(`<tileset [^>]*tilewidth="(\d+)"`)
	reTileH   = regexp.MustCompile(`<tileset [^>]*tileheight="(\d+)"`)
	reColumns = regexp.MustCompile(`<tileset [^>]*columns="(\d+)"`)
	reImage   = regexp.MustCompile(`<image source="([^"]+)" width="(\d+)" height="(\d+)"/>`)
)

func parseHeader(tsx string) (tsxHeader, error) {
	var h tsxHeader
	for _, f := range []struct {
		re  *regexp.Regexp
		out *int
	}{{reTileW, &h.tileW}, {reTileH, &h.tileH}, {reColumns, &h.columns}} {
		m := f.re.FindStringSubmatch(tsx)
		if m == nil {
			return h, fmt.Errorf("tileset header: missing %s", f.re)
		}
		*f.out, _ = strconv.Atoi(m[1])
	}
	m := reImage.FindStringSubmatch(tsx)
	if m == nil {
		return h, fmt.Errorf("tileset: missing <image>")
	}
	h.imageSource = m[1]
	return h, nil
}

// Validate checks the spec.
func (a *Anim) Validate() []string {
	var errs []string
	s := a.Spec
	if s.Tileset == "" || s.SourceRows <= 0 || s.Ms <= 0 {
		errs = append(errs, "tileset, sourceRows and ms are required")
	}
	if len(s.Sources) == 0 {
		errs = append(errs, "need at least one source")
	}
	for si, src := range s.Sources {
		if len(src.Frames) < 2 {
			errs = append(errs, fmt.Sprintf("source %d: need at least 2 frames", si))
		}
		if len(src.Frames) != len(s.Sources[0].Frames) {
			errs = append(errs, fmt.Sprintf("source %d: every source needs the same frame count", si))
		}
		if len(src.Layout) == 0 || len(src.Layout[0]) == 0 {
			errs = append(errs, fmt.Sprintf("source %d: layout is required", si))
			continue
		}
		for i, row := range src.Layout {
			if len(row) != len(src.Layout[0]) {
				errs = append(errs, fmt.Sprintf("source %d: layout row %d has %d tiles, want %d", si, i, len(row), len(src.Layout[0])))
			}
		}
		inLayout := map[int]bool{}
		for _, row := range src.Layout {
			for _, id := range row {
				inLayout[id] = true
			}
		}
		for alias, from := range src.Aliases {
			if !inLayout[from] {
				errs = append(errs, fmt.Sprintf("source %d: alias %d -> %d, but %d is not in the layout", si, alias, from, from))
			}
		}
	}
	if _, err := rig.ParseHex(s.WaterColor); err != nil {
		errs = append(errs, "waterColor: "+err.Error())
	}
	if s.Band <= 0 {
		errs = append(errs, "band must be > 0")
	}
	for _, p := range s.ColorMap {
		for _, c := range []string{p.From, p.To} {
			if _, err := rig.ParseHex(c); err != nil {
				errs = append(errs, "colorMap: "+err.Error())
			}
		}
	}
	if len(s.ColorMap) == 0 {
		errs = append(errs, "colorMap is required (map the animation onto the tileset palette)")
	}
	return errs
}

func dist2(a, b rig.RGB) int {
	dr, dg, db := int(a.R)-int(b.R), int(a.G)-int(b.G), int(a.B)-int(b.B)
	return dr*dr + dg*dg + db*db
}

// mapColor returns the palette colour for an animation pixel.
func (a *Anim) mapColor(c rig.RGB, pairs [][2]rig.RGB) rig.RGB {
	best, bd := c, -1
	for _, p := range pairs {
		if d := dist2(c, p[0]); bd < 0 || d < bd {
			best, bd = p[1], d
		}
	}
	return best
}

// Build returns the regenerated tileset image and .tsx.
func (a *Anim) Build() ([]rig.Output, error) {
	if errs := a.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("tileanim invalid:\n  %s", strings.Join(errs, "\n  "))
	}
	s := a.Spec
	tsxPath := a.path(s.Tileset)
	raw, err := os.ReadFile(tsxPath)
	if err != nil {
		return nil, err
	}
	tsx := string(raw)
	h, err := parseHeader(tsx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tsxPath, err)
	}
	imgPath := filepath.Join(filepath.Dir(tsxPath), filepath.FromSlash(h.imageSource))
	full, err := rig.ReadPNG(imgPath)
	if err != nil {
		return nil, err
	}
	srcH := s.SourceRows * h.tileH
	if full.Bounds().Dy() < srcH {
		return nil, fmt.Errorf("%s is %dpx tall, sourceRows needs %d", imgPath, full.Bounds().Dy(), srcH)
	}
	srcTiles := s.SourceRows * h.columns
	water, _ := rig.ParseHex(s.WaterColor)
	var pairs [][2]rig.RGB
	for _, p := range s.ColorMap {
		from, _ := rig.ParseHex(p.From)
		to, _ := rig.ParseHex(p.To)
		pairs = append(pairs, [2]rig.RGB{from, to})
	}

	tileAt := func(id int) *image.NRGBA {
		t := image.NewNRGBA(image.Rect(0, 0, h.tileW, h.tileH))
		ox, oy := (id%h.columns)*h.tileW, (id/h.columns)*h.tileH
		for y := 0; y < h.tileH; y++ {
			for x := 0; x < h.tileW; x++ {
				t.SetNRGBA(x, y, full.NRGBAAt(ox+x, oy+y))
			}
		}
		return t
	}

	// Where each tile's animation comes from: a source's frames and a cell.
	type origin struct {
		frames   []*image.NRGBA
		row, col int
	}
	origins := map[int]origin{}
	for si, src := range s.Sources {
		frames := make([]*image.NRGBA, len(src.Frames))
		wantW, wantH := len(src.Layout[0])*h.tileW, len(src.Layout)*h.tileH
		for i, f := range src.Frames {
			if frames[i], err = rig.ReadPNG(a.path(f)); err != nil {
				return nil, err
			}
			if b := frames[i].Bounds(); b.Dx() != wantW || b.Dy() != wantH {
				return nil, fmt.Errorf("source %d frame %s is %dx%d, layout needs %dx%d", si, f, b.Dx(), b.Dy(), wantW, wantH)
			}
		}
		cells := map[int]origin{}
		for r, row := range src.Layout {
			for c, id := range row {
				if id < 0 || id >= srcTiles {
					return nil, fmt.Errorf("source %d: layout tile %d is outside the %d source tiles", si, id, srcTiles)
				}
				if _, seen := cells[id]; !seen {
					cells[id] = origin{frames, r, c}
				}
			}
		}
		for alias, from := range src.Aliases {
			cells[alias] = cells[from]
		}
		for id, o := range cells {
			if _, taken := origins[id]; !taken {
				origins[id] = o
			}
		}
	}
	ids := make([]int, 0, len(origins))
	for id := range origins {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	frameCount := len(s.Sources[0].Frames)

	type animTile struct {
		id     int
		frames []*image.NRGBA
	}
	var animated []animTile
	for _, id := range ids {
		orig := tileAt(id)
		band := waterlineBand(orig, water, s.Band)
		if band == nil {
			continue // no waterline in this tile: it stays still
		}
		at := origins[id]
		var out []*image.NRGBA
		for _, fr := range at.frames {
			t := image.NewNRGBA(orig.Bounds())
			copy(t.Pix, orig.Pix)
			for y := 0; y < h.tileH; y++ {
				for x := 0; x < h.tileW; x++ {
					if !band[y*h.tileW+x] {
						continue
					}
					p := fr.NRGBAAt(at.col*h.tileW+x, at.row*h.tileH+y)
					c := a.mapColor(rig.RGB{R: p.R, G: p.G, B: p.B}, pairs)
					if orig.NRGBAAt(x, y).A == 0 && c == water {
						continue // see-through tile: the water layer below shows
					}
					t.SetNRGBA(x, y, color.NRGBA{c.R, c.G, c.B, 255})
				}
			}
			out = append(out, t)
		}
		animated = append(animated, animTile{id, out})
	}

	extra := len(animated) * frameCount
	rows := s.SourceRows + (extra+h.columns-1)/h.columns
	out := image.NewNRGBA(image.Rect(0, 0, h.columns*h.tileW, rows*h.tileH))
	for y := 0; y < srcH; y++ {
		for x := 0; x < h.columns*h.tileW; x++ {
			out.SetNRGBA(x, y, full.NRGBAAt(x, y))
		}
	}
	frameIDs := map[int][]int{}
	next := srcTiles
	for _, at := range animated {
		for _, fr := range at.frames {
			ox, oy := (next%h.columns)*h.tileW, (next/h.columns)*h.tileH
			for y := 0; y < h.tileH; y++ {
				for x := 0; x < h.tileW; x++ {
					out.SetNRGBA(ox+x, oy+y, fr.NRGBAAt(x, y))
				}
			}
			frameIDs[at.id] = append(frameIDs[at.id], next)
			next++
		}
	}

	needsWater := map[int]bool{}
	for _, at := range animated {
		orig := tileAt(at.id)
		for i := 3; i < len(orig.Pix); i += 4 {
			if orig.Pix[i] == 0 {
				needsWater[at.id] = true
				break
			}
		}
	}
	var waterTiles []int
	for id := 0; id < srcTiles; id++ {
		t := tileAt(id)
		all := true
		for i := 0; i < len(t.Pix) && all; i += 4 {
			all = t.Pix[i+3] == 255 && t.Pix[i] == water.R && t.Pix[i+1] == water.G && t.Pix[i+2] == water.B
		}
		if all {
			waterTiles = append(waterTiles, id)
		}
	}

	newTsx, err := a.patchTsx(tsx, srcTiles, rows*h.columns, rows*h.tileH, frameIDs, needsWater, waterTiles)
	if err != nil {
		return nil, err
	}
	return []rig.Output{
		{Path: imgPath, Data: rig.EncodePNG(out)},
		{Path: tsxPath, Data: []byte(newTsx)},
	}, nil
}

// waterlineBand marks pixels within dist (Chebyshev) of the tile's
// waterline: any pixel where water meets non-water. nil = no waterline.
func waterlineBand(t *image.NRGBA, water rig.RGB, dist int) []bool {
	w, h := t.Bounds().Dx(), t.Bounds().Dy()
	// Water is the water colour, or see-through (sand layered over a water
	// layer, as the world generator draws shores).
	isWater := func(x, y int) bool {
		c := t.NRGBAAt(x, y)
		return c.A == 0 || (c.R == water.R && c.G == water.G && c.B == water.B)
	}
	var line [][2]int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			for _, d := range [][2]int{{1, 0}, {0, 1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < w && ny < h && isWater(x, y) != isWater(nx, ny) {
					line = append(line, [2]int{x, y}, [2]int{nx, ny})
				}
			}
		}
	}
	if len(line) == 0 {
		return nil
	}
	band := make([]bool, w*h)
	for _, p := range line {
		for y := p[1] - dist; y <= p[1]+dist; y++ {
			for x := p[0] - dist; x <= p[0]+dist; x++ {
				if x >= 0 && y >= 0 && x < w && y < h {
					band[y*w+x] = true
				}
			}
		}
	}
	return band
}

const (
	syncedProp = `   <property name="synced" type="bool" value="true"/>`
	// waterProp marks tiles that are all water, so the server knows what
	// lies under a see-through shore tile.
	waterProp = `   <property name="water" type="bool" value="true"/>`
	// needsWaterProp marks see-through shore tiles: their tide only plays
	// where water is underneath (the same sand edge is also used over grass).
	needsWaterProp = `   <property name="needs_water_below" type="bool" value="true"/>`
)

var generatedProps = []string{syncedProp, waterProp, needsWaterProp}

var (
	reTileCount  = regexp.MustCompile(`(<tileset [^>]*tilecount=")\d+(")`)
	reImageH     = regexp.MustCompile(`(<image source="[^"]+" width="\d+" height=")\d+("/>)`)
	reAnimation  = regexp.MustCompile(`(?s)\n  <animation>.*?\n  </animation>`)
	reEmptyProps = regexp.MustCompile(`\n  <properties>\n  </properties>`)
	reEmptyTile  = regexp.MustCompile(`(<tile id="\d+"[^>/]*)>\n </tile>`)
)

// patchTsx removes what a previous run added, then adds animations for the
// current frames. Hand-written content (wangsets, collision, probability,
// other properties) is left as it is.
func (a *Anim) patchTsx(tsx string, srcTiles, tileCount, imageH int, frames map[int][]int, needsWater map[int]bool, waterTiles []int) (string, error) {
	// 1. Strip previous generated content.
	tsx = reAnimation.ReplaceAllStringFunc(tsx, func(block string) string {
		for _, m := range regexp.MustCompile(`tileid="(\d+)"`).FindAllStringSubmatch(block, -1) {
			if id, _ := strconv.Atoi(m[1]); id >= srcTiles {
				return "" // references generated frames: ours
			}
		}
		return block
	})
	for _, prop := range generatedProps {
		tsx = strings.ReplaceAll(tsx, "\n"+prop, "")
	}
	tsx = reEmptyProps.ReplaceAllString(tsx, "")
	tsx = reEmptyTile.ReplaceAllString(tsx, "$1/>")

	// 2. Header counts.
	tsx = reTileCount.ReplaceAllString(tsx, "${1}"+strconv.Itoa(tileCount)+"${2}")
	tsx = reImageH.ReplaceAllString(tsx, "${1}"+strconv.Itoa(imageH)+"${2}")

	// 3. Add animations (and their properties), then water markers.
	ids := make([]int, 0, len(frames))
	for id := range frames {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	var err error
	for _, id := range ids {
		var anim strings.Builder
		anim.WriteString("\n  <animation>")
		for _, fid := range frames[id] {
			fmt.Fprintf(&anim, "\n   <frame tileid=\"%d\" duration=\"%d\"/>", fid, a.Spec.Ms)
		}
		anim.WriteString("\n  </animation>")
		var props []string
		if a.synced() {
			props = append(props, syncedProp)
		}
		if needsWater[id] {
			props = append(props, needsWaterProp)
		}
		if tsx, err = addToTile(tsx, id, props, anim.String()); err != nil {
			return "", err
		}
	}
	for _, id := range waterTiles {
		if tsx, err = addToTile(tsx, id, []string{waterProp}, ""); err != nil {
			return "", err
		}
	}
	return tsx, nil
}

// addToTile adds generated properties and (optionally) an animation to tile
// id, creating the element body if the tileset has a self-closing one.
func addToTile(tsx string, id int, props []string, anim string) (string, error) {
	propBlock := func() string {
		if len(props) == 0 {
			return ""
		}
		return "\n  <properties>\n" + strings.Join(props, "\n") + "\n  </properties>"
	}
	reSelf := regexp.MustCompile(`\n <tile id="` + strconv.Itoa(id) + `"([^>]*?)/>`)
	if m := reSelf.FindStringSubmatchIndex(tsx); m != nil {
		attrs := tsx[m[2]:m[3]]
		repl := "\n <tile id=\"" + strconv.Itoa(id) + "\"" + attrs + ">" + propBlock() + anim + "\n </tile>"
		return tsx[:m[0]] + repl + tsx[m[1]:], nil
	}
	reOpen := regexp.MustCompile(`\n <tile id="` + strconv.Itoa(id) + `"[^>]*>`)
	m := reOpen.FindStringIndex(tsx)
	if m == nil {
		return "", fmt.Errorf("tileset has no <tile id=\"%d\"> element", id)
	}
	end := strings.Index(tsx[m[1]:], "\n </tile>")
	if end < 0 {
		return "", fmt.Errorf("tile %d is not closed", id)
	}
	end += m[1]
	body := tsx[m[1]:end]
	if len(props) > 0 {
		if strings.Contains(body, "\n  <properties>") {
			body = strings.Replace(body, "\n  <properties>", "\n  <properties>\n"+strings.Join(props, "\n"), 1)
		} else {
			body = propBlock() + body
		}
	}
	return tsx[:m[1]] + body + anim + tsx[end:], nil
}

// AnimatedCount is how many tiles a build animates (for reporting).
func AnimatedCount(outs []rig.Output) int {
	for _, o := range outs {
		if strings.HasSuffix(o.Path, ".tsx") {
			return strings.Count(string(o.Data), "<animation>")
		}
	}
	return 0
}
