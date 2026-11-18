package rig

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ImportOptions configures the one-time migration from a baked sheet.
type ImportOptions struct {
	Sheet  string // baked spritesheet PNG
	Tsx    string // Tiled tileset with <animation>s naming the clips
	OutDir string // rig directory to create
	Name   string // entity / tileset name, e.g. "player"
	// Outputs written into rig.yml (relative to OutDir).
	Outputs Outputs
	Force   bool
}

// Colour classes found in the baked sheet.
const (
	clsNone = iota
	clsOutline
	clsSkin
	clsHair
	clsShirt
	clsPants
)

var classSlot = map[int]string{clsSkin: "skin", clsHair: "hair", clsShirt: "shirt", clsPants: "pants"}

var rampSize = map[int]int{clsOutline: 2, clsSkin: 4, clsHair: 3, clsShirt: 4, clsPants: 3}

// Import slices a baked sheet into a rig. It is the only code that reads the
// legacy sheet; after it runs, the rig directory is the source of truth.
func Import(o ImportOptions) error {
	if entries, _ := os.ReadDir(o.OutDir); len(entries) > 0 && !o.Force {
		return fmt.Errorf("%s is not empty (use -force to overwrite)", o.OutDir)
	}
	sheet, err := ReadPNG(o.Sheet)
	if err != nil {
		return err
	}
	clips, err := readTsxClips(o.Tsx, o.Name)
	if err != nil {
		return err
	}
	const fs = 16
	cols := sheet.Bounds().Dx() / fs
	rows := sheet.Bounds().Dy() / fs
	nFrames := cols * rows

	// Name each frame after the clip that first uses it.
	frameName := make([]string, nFrames)
	clipNames := sortedKeys(clips)
	for _, cn := range clipNames {
		for i, st := range clips[cn] {
			if st.id < nFrames && frameName[st.id] == "" {
				frameName[st.id] = fmt.Sprintf("%s_%d", cn, i)
			}
		}
	}
	for i := range frameName {
		if frameName[i] == "" {
			frameName[i] = fmt.Sprintf("frame_%d", i)
		}
	}

	// 1. Classify every opaque pixel.
	type px struct {
		c   RGB
		cls int
	}
	frames := make([][fs][fs]px, nFrames)
	for f := 0; f < nFrames; f++ {
		ox, oy := (f%cols)*fs, (f/cols)*fs
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				if c, ok := rgbAt(sheet, ox+x, oy+y); ok {
					frames[f][y][x] = px{c, classify(c, y)}
				}
			}
		}
		// Ambiguous darks take the majority class of their neighbours.
		var fixed [fs][fs]int
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				p := frames[f][y][x]
				fixed[y][x] = p.cls
				if p.cls != clsHair && p.cls != clsPants {
					continue
				}
				if _, _, l := p.c.HSL(); l >= 0.22 {
					continue
				}
				counts := map[int]int{}
				for _, d := range neighbours8 {
					nx, ny := x+d[0], y+d[1]
					if nx < 0 || ny < 0 || nx >= fs || ny >= fs {
						continue
					}
					if c := frames[f][ny][nx].cls; c != clsNone && c != clsOutline {
						counts[c]++
					}
				}
				if best := argmax(counts, []int{clsShirt, clsHair, clsPants, clsSkin}); best != clsNone && counts[best] > counts[p.cls] {
					fixed[y][x] = best
				}
			}
		}
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				frames[f][y][x].cls = fixed[y][x]
			}
		}
	}

	// 2. Build ramps per class from the colours actually used.
	samples := map[int][]RGB{}
	for f := range frames {
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				if p := frames[f][y][x]; p.cls != clsNone {
					samples[p.cls] = append(samples[p.cls], p.c)
				}
			}
		}
	}
	ramps := map[int][]RGB{}
	used := map[RGB]bool{}
	for _, cls := range []int{clsOutline, clsSkin, clsHair, clsShirt, clsPants} {
		ramp := buildRamp(samples[cls], rampSize[cls])
		for i := range ramp {
			for used[ramp[i]] { // key colours must be unique across slots
				ramp[i].B++
			}
			used[ramp[i]] = true
		}
		ramps[cls] = ramp
	}

	// 3. Quantise and split into layers.
	layers := []string{"pants", "body", "shirt", "head", "hair"}
	parts := map[string]map[[32]byte]string{} // layer -> hash -> part name
	partImgs := map[string]map[string]*image.NRGBA{}
	for _, l := range layers {
		parts[l] = map[[32]byte]string{}
		partImgs[l] = map[string]*image.NRGBA{}
	}
	var poses []Pose
	for f := 0; f < nFrames; f++ {
		neck := neckRow(func(x, y int) bool { return frames[f][y][x].cls != clsNone })
		layerOf := [fs][fs]string{}
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				switch frames[f][y][x].cls {
				case clsSkin:
					if y < neck {
						layerOf[y][x] = "head"
					} else {
						layerOf[y][x] = "body"
					}
				case clsHair:
					layerOf[y][x] = "hair"
				case clsShirt:
					layerOf[y][x] = "shirt"
				case clsPants:
					layerOf[y][x] = "pants"
				}
			}
		}
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				if frames[f][y][x].cls != clsOutline {
					continue
				}
				counts := map[string]int{}
				for _, d := range neighbours8 {
					nx, ny := x+d[0], y+d[1]
					if nx < 0 || ny < 0 || nx >= fs || ny >= fs {
						continue
					}
					if l := layerOf[ny][nx]; l != "" && frames[f][ny][nx].cls != clsOutline {
						counts[l]++
					}
				}
				best, n := "", 0
				for _, l := range []string{"hair", "head", "shirt", "body", "pants"} {
					if counts[l] > n {
						best, n = l, counts[l]
					}
				}
				if best == "" {
					best = "body"
					if y < neck {
						best = "head"
					}
				}
				layerOf[y][x] = best
			}
		}

		pose := Pose{Name: frameName[f], Parts: map[string]string{}}
		for _, l := range layers {
			img := image.NewNRGBA(image.Rect(0, 0, fs, fs))
			empty := true
			for y := 0; y < fs; y++ {
				for x := 0; x < fs; x++ {
					p := frames[f][y][x]
					if layerOf[y][x] != l {
						continue
					}
					setRGB(img, x, y, nearestByLuma(p.c, ramps[p.cls]))
					empty = false
				}
			}
			if empty {
				continue
			}
			h := sha256.Sum256(img.Pix)
			name, ok := parts[l][h]
			if !ok {
				name = frameName[f]
				parts[l][h] = name
				partImgs[l][name] = img
			}
			pose.Parts[l] = name
		}
		poses = append(poses, pose)
	}

	// 4. Write the rig.
	if err := os.MkdirAll(o.OutDir, 0o755); err != nil {
		return err
	}
	variant := map[string]string{"hair": "short"}
	for _, l := range layers {
		v := variant[l]
		if v == "" {
			v = "default"
		}
		dir := filepath.Join(o.OutDir, "parts", l, v)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for name, img := range partImgs[l] {
			if err := os.WriteFile(filepath.Join(dir, name+".png"), EncodePNG(img), 0o644); err != nil {
				return err
			}
		}
	}

	spec := RigSpec{
		Name:      o.Name,
		FrameSize: fs,
		Columns:   cols,
		Slots:     map[string][]string{},
		SlotOrder: []string{"skin", "hair", "shirt", "pants"},
		Palettes:  map[string]map[string][]string{},
		Defaults:  Defaults{Palettes: map[string]string{}, Parts: map[string]string{"hair": "short"}},
		Outputs:   o.Outputs,
	}
	for _, l := range layers {
		ls := LayerSpec{Name: l}
		if v, ok := variant[l]; ok {
			ls.Variants = []string{v}
		}
		spec.Layers = append(spec.Layers, ls)
	}
	for cls, slot := range classSlot {
		hexes := hexList(ramps[cls])
		spec.Slots[slot] = hexes
		spec.Palettes[slot] = map[string][]string{"default": hexes}
		spec.Defaults.Palettes[slot] = "default"
	}
	spec.Fixed = hexList(ramps[clsOutline])

	cs := ClipsSpec{Clips: map[string][]Step{}, States: map[string]string{}, InitialState: "idle"}
	for _, cn := range clipNames {
		for _, st := range clips[cn] {
			cs.Clips[cn] = append(cs.Clips[cn], Step{Pose: frameName[st.id], Ms: st.ms})
		}
		cs.States[cn] = cn
	}
	cs.Idle = IdleSpec{Base: "idle"}

	header := "# Generated once by `spritegen import`; now the source of truth. Edit freely.\n"
	if err := writeYAML(filepath.Join(o.OutDir, "rig.yml"), header, spec); err != nil {
		return err
	}
	if err := writeYAML(filepath.Join(o.OutDir, "poses.yml"), header, posesFile{Poses: poses}); err != nil {
		return err
	}
	return writeYAML(filepath.Join(o.OutDir, "clips.yml"), header, cs)
}

var neighbours8 = [][2]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}}

func classify(c RGB, y int) int {
	h, s, l := c.HSL()
	switch {
	case l < 0.13:
		return clsOutline
	case l >= 0.62 && (h >= 320 || h <= 50 || s < 0.25):
		return clsSkin
	case s >= 0.35 && h >= 270 && h < 345:
		return clsShirt
	case l >= 0.55:
		return clsSkin
	case y >= 12:
		return clsPants
	default:
		return clsHair
	}
}

func argmax(counts map[int]int, order []int) int {
	best, n := clsNone, 0
	for _, c := range order {
		if counts[c] > n {
			best, n = c, counts[c]
		}
	}
	return best
}

// neckRow is the narrowest opaque row between y=5 and y=8: rows above it are
// the head.
func neckRow(opaque func(x, y int) bool) int {
	best, bestN := 6, 1<<30
	for y := 5; y <= 8; y++ {
		n := 0
		for x := 0; x < 16; x++ {
			if opaque(x, y) {
				n++
			}
		}
		if n > 0 && n < bestN {
			best, bestN = y, n
		}
	}
	return best
}

// buildRamp splits samples into n equal-count luma bands and averages each.
func buildRamp(samples []RGB, n int) []RGB {
	if len(samples) == 0 {
		return nil
	}
	s := append([]RGB(nil), samples...)
	sort.SliceStable(s, func(i, j int) bool { return s[i].Luma() < s[j].Luma() })
	ramp := make([]RGB, 0, n)
	for b := 0; b < n; b++ {
		lo, hi := b*len(s)/n, (b+1)*len(s)/n
		if hi <= lo {
			continue
		}
		var r, g, bl int
		for _, c := range s[lo:hi] {
			r += int(c.R)
			g += int(c.G)
			bl += int(c.B)
		}
		k := hi - lo
		ramp = append(ramp, RGB{uint8(r / k), uint8(g / k), uint8(bl / k)})
	}
	return ramp
}

func nearestByLuma(c RGB, ramp []RGB) RGB {
	best, bd := ramp[0], 1e9
	for _, r := range ramp {
		d := c.Luma() - r.Luma()
		if d < 0 {
			d = -d
		}
		if d < bd {
			best, bd = r, d
		}
	}
	return best
}

func hexList(cs []RGB) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Hex()
	}
	return out
}

type tsxStep struct{ id, ms int }

func readTsxClips(path, name string) (map[string][]tsxStep, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ts struct {
		Tiles []struct {
			ID        int    `xml:"id,attr"`
			Type      string `xml:"type,attr"`
			Animation struct {
				Frames []struct {
					TileID   int `xml:"tileid,attr"`
					Duration int `xml:"duration,attr"`
				} `xml:"frame"`
			} `xml:"animation"`
		} `xml:"tile"`
	}
	if err := xml.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	clips := map[string][]tsxStep{}
	for _, t := range ts.Tiles {
		clip := strings.TrimPrefix(t.Type, name+"_")
		if t.Type == name {
			clip = "idle"
		}
		for _, f := range t.Animation.Frames {
			ms := f.Duration
			if ms <= 0 {
				ms = 100
			}
			clips[clip] = append(clips[clip], tsxStep{f.TileID, ms})
		}
	}
	return clips, nil
}

func writeYAML(path, header string, v any) error {
	var b strings.Builder
	b.WriteString(header)
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
