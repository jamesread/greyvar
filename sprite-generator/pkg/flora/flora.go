// Package flora builds an animated decoration tileset (flowers and similar)
// from small transparent parts plus flora.yml. Each flower becomes one
// animated Tiled tile whose frames sway its "heads" sideways.
//
//	res/sprites/flora/
//	  flora.yml        tileset name, outputs, sway pattern, one entry per flower
//	  parts/<name>.png 16x16 transparent flower, drawn upright (frame 0)
//
// `spritegen flora build` writes the tileset PNG and .tsx; never edit those.
package flora

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
	"gopkg.in/yaml.v3"
)

// Spec is flora.yml.
type Spec struct {
	Name     string  `yaml:"name"`
	TileSize int     `yaml:"tileSize"`
	Columns  int     `yaml:"columns"`
	Outputs  Outputs `yaml:"outputs"`
	Sway     Sway    `yaml:"sway"`
	Flora    []Entry `yaml:"flora"`
	// Migrate configures `flora migrate` (moving baked flower tiles to a decor layer).
	Migrate Migrate `yaml:"migrate"`
	// Decor marks every tile decor=true (default true): drawn on top, never
	// collides. Set false for tiles with their own rules, e.g. crops.
	Decor *bool `yaml:"decor,omitempty"`
}

// IsDecor reports whether tiles get the decor=true property.
func (s Spec) IsDecor() bool { return s.Decor == nil || *s.Decor }

// Migrate maps a source tileset (basename, e.g. grass_terrains.tsx) to the
// plain tile drawn where a baked flower used to be.
type Migrate struct {
	Replacement map[string]int `yaml:"replacement"`
}

// Outputs are relative to the flora directory.
type Outputs struct {
	Image string `yaml:"image"`
	Tsx   string `yaml:"tsx"`
}

// Sway is the shared animation: a sideways offset (px) and duration per frame.
type Sway struct {
	Offsets []int `yaml:"offsets"`
	Ms      []int `yaml:"ms"`
}

// Entry is one flower tile.
type Entry struct {
	Name string `yaml:"name"`
	Part string `yaml:"part"`
	// Sources are baked tiles showing this flower (e.g.
	// grass_terrains.tsx#126); `flora migrate` replaces them. Optional.
	Sources []string `yaml:"sources,omitempty"`
	// Heads are inclusive boxes [x0, y0, x1, y1] whose pixels sway.
	Heads [][4]int `yaml:"heads"`
	// Phases shift each head through the sway pattern so clustered flowers
	// do not move together (default 0 for every head).
	Phases []int `yaml:"phases,omitempty"`
	// Props are extra Tiled tile properties, e.g. {crop: hay, stage: "0"}.
	Props map[string]string `yaml:"props,omitempty"`
}

// Flora is a loaded flora directory.
type Flora struct {
	Dir  string
	Spec Spec
}

// Load reads flora.yml.
func Load(dir string) (*Flora, error) {
	data, err := os.ReadFile(filepath.Join(dir, "flora.yml"))
	if err != nil {
		return nil, err
	}
	f := &Flora{Dir: dir}
	if err := yaml.Unmarshal(data, &f.Spec); err != nil {
		return nil, fmt.Errorf("flora.yml: %w", err)
	}
	if f.Spec.TileSize == 0 {
		f.Spec.TileSize = 16
	}
	if f.Spec.Columns == 0 {
		f.Spec.Columns = 8
	}
	return f, nil
}

func (f *Flora) path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Clean(filepath.Join(f.Dir, rel))
}

func (f *Flora) partPath(e Entry) string {
	part := e.Part
	if part == "" {
		part = e.Name + ".png"
	}
	return filepath.Join(f.Dir, "parts", part)
}

// FrameCount is the number of animation frames per flower.
func (f *Flora) FrameCount() int { return len(f.Spec.Sway.Offsets) }

// Validate returns problems; empty means valid.
func (f *Flora) Validate() []string {
	var errs []string
	add := func(format string, a ...any) { errs = append(errs, fmt.Sprintf(format, a...)) }
	s := f.Spec
	ts := s.TileSize
	if s.Name == "" || s.Outputs.Image == "" || s.Outputs.Tsx == "" {
		add("flora.yml: name, outputs.image and outputs.tsx are required")
	}
	if len(s.Sway.Offsets) < 2 || len(s.Sway.Offsets) != len(s.Sway.Ms) {
		add("flora.yml: sway.offsets and sway.ms need the same length, at least 2")
	}
	for i, ms := range s.Sway.Ms {
		if ms <= 0 {
			add("flora.yml: sway.ms[%d] must be > 0", i)
		}
	}
	seen := map[string]bool{}
	for _, e := range s.Flora {
		if e.Name == "" || seen[e.Name] {
			add("flora entry %q: names must be unique and non-empty", e.Name)
		}
		seen[e.Name] = true
		img, err := rig.ReadPNG(f.partPath(e))
		if err != nil {
			add("flora %s: %v", e.Name, err)
			continue
		}
		if img.Bounds().Dx() != ts || img.Bounds().Dy() != ts {
			add("flora %s: part is %dx%d, want %dx%d", e.Name, img.Bounds().Dx(), img.Bounds().Dy(), ts, ts)
			continue
		}
		for y := 0; y < ts; y++ {
			for x := 0; x < ts; x++ {
				if a := img.NRGBAAt(x, y).A; a != 0 && a != 255 {
					add("flora %s: pixel (%d,%d) is semi-transparent", e.Name, x, y)
				}
			}
		}
		if len(e.Heads) == 0 {
			add("flora %s: needs at least one head box", e.Name)
		}
		if len(e.Phases) != 0 && len(e.Phases) != len(e.Heads) {
			add("flora %s: phases must have one entry per head", e.Name)
		}
		for h, b := range e.Heads {
			if b[0] > b[2] || b[1] > b[3] || b[0] < 0 || b[1] < 0 || b[2] >= ts || b[3] >= ts {
				add("flora %s head %d: box %v outside the %dx%d tile", e.Name, h, b, ts, ts)
				continue
			}
			for _, off := range s.Sway.Offsets {
				for y := b[1]; y <= b[3]; y++ {
					for x := b[0]; x <= b[2]; x++ {
						if img.NRGBAAt(x, y).A != 0 && (x+off < 0 || x+off >= ts) {
							add("flora %s head %d: offset %d pushes pixel (%d,%d) out of the tile", e.Name, h, off, x, y)
						}
					}
				}
			}
		}
	}
	return errs
}

func inBox(b [4]int, x, y int) bool { return x >= b[0] && x <= b[2] && y >= b[1] && y <= b[3] }

// frame renders flower e at animation frame n.
func (f *Flora) frame(e Entry, part *image.NRGBA, n int) *image.NRGBA {
	ts := f.Spec.TileSize
	out := image.NewNRGBA(image.Rect(0, 0, ts, ts))
	offs := f.Spec.Sway.Offsets
	head := func(x, y int) int {
		for h, b := range e.Heads {
			if inBox(b, x, y) {
				return h
			}
		}
		return -1
	}
	// Static pixels first, then swaying heads on top.
	for pass := 0; pass < 2; pass++ {
		for y := 0; y < ts; y++ {
			for x := 0; x < ts; x++ {
				c := part.NRGBAAt(x, y)
				if c.A == 0 {
					continue
				}
				h := head(x, y)
				if (pass == 0) != (h < 0) {
					continue
				}
				dx := 0
				if h >= 0 {
					phase := 0
					if h < len(e.Phases) {
						phase = e.Phases[h]
					}
					dx = offs[((n+phase)%len(offs)+len(offs))%len(offs)]
				}
				if x+dx >= 0 && x+dx < ts {
					out.SetNRGBA(x+dx, y, c)
				}
			}
		}
	}
	return out
}

// Build returns the generated tileset image and .tsx.
func (f *Flora) Build() ([]rig.Output, error) {
	if errs := f.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("flora invalid:\n  %s", strings.Join(errs, "\n  "))
	}
	s := f.Spec
	ts, cols, n := s.TileSize, s.Columns, f.FrameCount()
	tiles := len(s.Flora) * n
	rows := (tiles + cols - 1) / cols
	sheet := image.NewNRGBA(image.Rect(0, 0, cols*ts, rows*ts))
	for i, e := range s.Flora {
		part, err := rig.ReadPNG(f.partPath(e))
		if err != nil {
			return nil, err
		}
		for fr := 0; fr < n; fr++ {
			id := i*n + fr
			img := f.frame(e, part, fr)
			ox, oy := (id%cols)*ts, (id/cols)*ts
			for y := 0; y < ts; y++ {
				for x := 0; x < ts; x++ {
					if c := img.NRGBAAt(x, y); c.A != 0 {
						sheet.SetNRGBA(ox+x, oy+y, c)
					}
				}
			}
		}
	}

	imgPath, tsxPath := f.path(s.Outputs.Image), f.path(s.Outputs.Tsx)
	rel, err := filepath.Rel(filepath.Dir(tsxPath), imgPath)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, "<!-- GENERATED by sprite-generator from res/sprites/%s. Do not edit; run `make sprites`. -->\n", s.Name)
	fmt.Fprintf(&b, `<tileset version="1.10" tiledversion="1.12.2" name="%s" tilewidth="%d" tileheight="%d" tilecount="%d" columns="%d">`+"\n",
		s.Name, ts, ts, tiles, cols)
	fmt.Fprintf(&b, ` <image source="%s" width="%d" height="%d"/>`+"\n", filepath.ToSlash(rel), cols*ts, rows*ts)
	for i, e := range s.Flora {
		first := i * n
		fmt.Fprintf(&b, ` <tile id="%d" type="%s">`+"\n  <properties>\n", first, s.Name)
		if s.IsDecor() {
			b.WriteString(`   <property name="decor" type="bool" value="true"/>` + "\n")
		}
		fmt.Fprintf(&b, `   <property name="flora" value="%s"/>`+"\n", e.Name)
		for _, k := range sortedPropKeys(e.Props) {
			fmt.Fprintf(&b, `   <property name="%s" value="%s"/>`+"\n", k, e.Props[k])
		}
		if len(e.Sources) > 0 {
			fmt.Fprintf(&b, `   <property name="sources" value="%s"/>`+"\n", strings.Join(e.Sources, ","))
		}
		b.WriteString("  </properties>\n  <animation>\n")
		for fr := 0; fr < n; fr++ {
			fmt.Fprintf(&b, `   <frame tileid="%d" duration="%d"/>`+"\n", first+fr, s.Sway.Ms[fr])
		}
		b.WriteString("  </animation>\n </tile>\n")
	}
	b.WriteString("</tileset>\n")

	return []rig.Output{
		{Path: imgPath, Data: rig.EncodePNG(sheet)},
		{Path: tsxPath, Data: []byte(b.String())},
	}, nil
}

// Preview renders every flower's frames (rows = flowers, columns = frames),
// upscaled on a grass-green background, for visual review.
func (f *Flora) Preview(scale int) (*image.NRGBA, error) {
	if scale <= 0 {
		scale = 8
	}
	ts, n := f.Spec.TileSize, f.FrameCount()
	cell := (ts + 1) * scale
	out := image.NewNRGBA(image.Rect(0, 0, n*cell, len(f.Spec.Flora)*cell))
	bg := [3]uint8{0xb5, 0xe6, 0x1d}
	for i := range out.Pix {
		if i%4 == 3 {
			out.Pix[i] = 255
		} else {
			out.Pix[i] = bg[i%4]
		}
	}
	for i, e := range f.Spec.Flora {
		part, err := rig.ReadPNG(f.partPath(e))
		if err != nil {
			return nil, err
		}
		for fr := 0; fr < n; fr++ {
			img := f.frame(e, part, fr)
			for y := 0; y < ts; y++ {
				for x := 0; x < ts; x++ {
					c := img.NRGBAAt(x, y)
					if c.A == 0 {
						continue
					}
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							out.SetNRGBA(fr*cell+x*scale+dx, i*cell+y*scale+dy, c)
						}
					}
				}
			}
		}
	}
	return out, nil
}

func sortedPropKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
