package rig

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Outfits are one-piece garments (dresses, ...) on their own layer, drawn
// above the shirt and trousers. Each outfit part covers every pixel the
// shirt and trousers draw in its poses, so whatever is worn underneath never
// shows. Parts are derived from the shirt and trouser parts (with their pose
// offsets), so sleeves, arms and walking legs line up in every frame.
//
// A pose names its outfit part after the shirt/trousers it covers:
// OutfitPartName. `part outfit` fills poses.yml's outfit entries and writes
// the parts for a variant: "none" (empty: the default, nothing drawn) or a
// style such as "dress".

// OutfitOptions configures the outfit op.
type OutfitOptions struct {
	Layer        string // outfit layer (default "outfit")
	Shirt, Pants string // layers covered (default "shirt", "pants")
	From         string // variant of the covered layers to derive from (default "default")
	To           string // variant to write
	Style        string // "none" or "dress"
	Garment      string // main colour slot (default "garment")
	Trim         string // trim colour slot (default "trim")
	SkirtRows    int    // skirt height below the waist, incl. the hem (default 3)
	Force        bool
}

// OutfitPartName is the part a pose uses on the outfit layer: the shirt and
// trousers parts it covers, plus their offsets when they are moved.
func OutfitPartName(p Pose, shirt, pants string) string {
	name := p.Parts[shirt] + "__" + p.Parts[pants]
	so, po := p.Offsets[shirt], p.Offsets[pants]
	if so != [2]int{} || po != [2]int{} {
		n := func(v int) string {
			if v < 0 {
				return fmt.Sprintf("m%d", -v)
			}
			return fmt.Sprint(v)
		}
		name += fmt.Sprintf("__s%s_%s_p%s_%s", n(so[0]), n(so[1]), n(po[0]), n(po[1]))
	}
	return name
}

func (o *OutfitOptions) defaults() {
	if o.Layer == "" {
		o.Layer = "outfit"
	}
	if o.Shirt == "" {
		o.Shirt = "shirt"
	}
	if o.Pants == "" {
		o.Pants = "pants"
	}
	if o.From == "" {
		o.From = "default"
	}
	if o.Garment == "" {
		o.Garment = "garment"
	}
	if o.Trim == "" {
		o.Trim = "trim"
	}
	if o.SkirtRows <= 0 {
		o.SkirtRows = 3
	}
	if o.Style == "" {
		o.Style = o.To
	}
}

// OutfitPoseParts returns, for every pose, the outfit part it should use.
func (r *Rig) OutfitPoseParts(o OutfitOptions) map[string]string {
	o.defaults()
	out := map[string]string{}
	for _, p := range r.Poses {
		if _, ok := p.Parts[o.Shirt]; !ok {
			continue
		}
		if _, ok := p.Parts[o.Pants]; !ok {
			continue
		}
		out[p.Name] = OutfitPartName(p, o.Shirt, o.Pants)
	}
	return out
}

// DeriveOutfit writes <layer>/<To>/<part>.png for every outfit part the poses
// use, and returns the written paths.
func (r *Rig) DeriveOutfit(o OutfitOptions) ([]string, error) {
	o.defaults()
	if o.To == "" {
		return nil, fmt.Errorf("outfit: -to is required")
	}
	fs := r.Spec.FrameSize
	parts := map[string]Pose{}
	for _, p := range r.Poses {
		name, ok := r.OutfitPoseParts(o)[p.Name]
		if !ok {
			continue
		}
		if have, ok := p.Parts[o.Layer]; ok && have != name {
			return nil, fmt.Errorf("pose %s uses outfit part %q, want %q (run `part outfit -poses`)", p.Name, have, name)
		}
		if _, seen := parts[name]; !seen {
			parts[name] = p
		}
	}
	names := make([]string, 0, len(parts))
	for n := range parts {
		names = append(names, n)
	}
	sort.Strings(names)

	var draw func(p Pose) (*image.NRGBA, error)
	switch o.Style {
	case "none":
		draw = func(Pose) (*image.NRGBA, error) { return image.NewNRGBA(image.Rect(0, 0, fs, fs)), nil }
	case "dress":
		d, err := r.newDresser(o)
		if err != nil {
			return nil, err
		}
		draw = d.draw
	default:
		return nil, fmt.Errorf("outfit: unknown style %q (none, dress)", o.Style)
	}

	var written []string
	for _, name := range names {
		img, err := draw(parts[name])
		if err != nil {
			return nil, err
		}
		out := r.PartPath(o.Layer, o.To, name)
		if _, err := os.Stat(out); err == nil && !o.Force {
			return nil, fmt.Errorf("%s exists (use -force)", out)
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(out, EncodePNG(img), 0o644); err != nil {
			return nil, err
		}
		written = append(written, out)
	}
	return written, nil
}

// SetOutfitPoses rewrites poses.yml so every pose names its outfit part.
// Each pose's `parts:` is a one-line flow map; the outfit entry is added (or
// replaced) at its end, leaving comments and layout alone.
func (r *Rig) SetOutfitPoses(o OutfitOptions) (int, error) {
	o.defaults()
	path := filepath.Join(r.Dir, "poses.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	want := r.OutfitPoseParts(o)
	lines := strings.Split(string(raw), "\n")
	pose, changed := "", 0
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "- name:") {
			pose = strings.TrimSpace(strings.TrimPrefix(t, "- name:"))
			continue
		}
		if !strings.HasPrefix(t, "parts:") || pose == "" {
			continue
		}
		name, ok := want[pose]
		if !ok {
			continue
		}
		open, close := strings.Index(line, "{"), strings.LastIndex(line, "}")
		if open < 0 || close < open {
			return 0, fmt.Errorf("poses.yml: pose %s: parts is not a one-line {...} map", pose)
		}
		var kept []string
		for _, kv := range strings.Split(line[open+1:close], ",") {
			if k, _, _ := strings.Cut(strings.TrimSpace(kv), ":"); strings.TrimSpace(k) != o.Layer && strings.TrimSpace(kv) != "" {
				kept = append(kept, strings.TrimSpace(kv))
			}
		}
		kept = append(kept, o.Layer+": "+name)
		next := line[:open+1] + strings.Join(kept, ", ") + line[close:]
		if next != line {
			lines[i] = next
			changed++
		}
	}
	if changed > 0 {
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
			return 0, err
		}
	}
	return changed, nil
}

// dresser draws the "dress" style: the shirt becomes the bodice (garment
// colours, a trim sash at the waist), the top SkirtRows of the trousers an
// A-line skirt (flaring 1px each side below the waist, shaded light left /
// dark right, a trim hem), and trousers below the hem bare legs (shoes and
// outline stay, so walk cycles still show the feet).
type dresser struct {
	r                   *Rig
	o                   OutfitOptions
	outline             RGB
	shirtIdx, pantsIdx  map[RGB]int
	garment, trim, skin []RGB
}

func (r *Rig) newDresser(o OutfitOptions) (*dresser, error) {
	outline, _, err := r.ColourRef("fixed.0")
	if err != nil {
		return nil, err
	}
	ramp := func(slot string) []RGB {
		var out []RGB
		for _, k := range r.Spec.Slots[slot] {
			c, _ := ParseHex(k)
			out = append(out, c)
		}
		return out
	}
	index := func(cs []RGB) map[RGB]int {
		m := map[RGB]int{}
		for i, c := range cs {
			m[c] = i
		}
		return m
	}
	d := &dresser{r: r, o: o, outline: outline,
		garment: ramp(o.Garment), trim: ramp(o.Trim), skin: ramp("skin")}
	d.shirtIdx, d.pantsIdx = index(ramp(o.Shirt)), index(ramp(o.Pants))
	if len(d.garment) < 4 || len(d.trim) < 3 || len(d.skin) < 2 {
		return nil, fmt.Errorf("dress needs colour slots %s (4 colours), %s (3) and skin", o.Garment, o.Trim)
	}
	return d, nil
}

func (d *dresser) layer(layer string, p Pose) (*image.NRGBA, error) {
	fs := d.r.Spec.FrameSize
	src, err := ReadPNG(d.r.PartPath(layer, d.o.From, p.Parts[layer]))
	if err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, fs, fs))
	off := p.Offsets[layer]
	for y := 0; y < fs; y++ {
		for x := 0; x < fs; x++ {
			if c, ok := rgbAt(src, x, y); ok {
				if nx, ny := x+off[0], y+off[1]; nx >= 0 && ny >= 0 && nx < fs && ny < fs {
					setRGB(out, nx, ny, c)
				}
			}
		}
	}
	return out, nil
}

func (d *dresser) draw(p Pose) (*image.NRGBA, error) {
	fs := d.r.Spec.FrameSize
	shirt, err := d.layer(d.o.Shirt, p)
	if err != nil {
		return nil, err
	}
	pants, err := d.layer(d.o.Pants, p)
	if err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, fs, fs))
	g, t := d.garment, d.trim

	// Bodice: the shirt in garment colours.
	shirtBottom := -1
	for y := 0; y < fs; y++ {
		for x := 0; x < fs; x++ {
			if c, ok := rgbAt(shirt, x, y); ok {
				if _, cloth := d.shirtIdx[c]; cloth {
					shirtBottom = max(shirtBottom, y)
				}
			}
		}
	}
	for y := 0; y < fs; y++ {
		for x := 0; x < fs; x++ {
			c, ok := rgbAt(shirt, x, y)
			if !ok {
				continue
			}
			if i, cloth := d.shirtIdx[c]; cloth {
				c = g[min(i, len(g)-1)]
			}
			setRGB(img, x, y, c)
		}
	}
	// Skirt: the top rows of the trousers, below the bodice.
	top := -1
	for y := 0; y < fs && top < 0; y++ {
		for x := 0; x < fs; x++ {
			if _, ok := rgbAt(pants, x, y); ok {
				top = y
				break
			}
		}
	}
	if top < 0 {
		return img, nil
	}
	// Waist: the rows between the bodice and the skirt (where the body
	// shows between shirt and trousers) are filled across the torso, and the
	// first of them is a trim sash; with no gap, the bodice's last row is.
	sl, sr := fs, -1
	for x := 0; x < fs; x++ {
		if c, ok := rgbAt(shirt, x, shirtBottom); ok && shirtBottom >= 0 {
			if _, cloth := d.shirtIdx[c]; cloth {
				sl, sr = min(sl, x), max(sr, x)
			}
		}
	}
	sash := shirtBottom
	if shirtBottom >= 0 && shirtBottom+1 < top && sr >= sl {
		sash = shirtBottom + 1
		for y := shirtBottom + 1; y < top; y++ {
			for x := sl - 1; x <= sr+1; x++ {
				if x < 0 || x >= fs {
					continue
				}
				c := g[2]
				if y == sash {
					c = t[1]
				}
				if x == sl-1 || x == sr+1 {
					c = d.outline
				}
				setRGB(img, x, y, c)
			}
		}
	} else if sash >= 0 {
		for x := sl; x <= sr; x++ {
			setRGB(img, x, sash, t[1])
		}
	}
	hem := min(top+d.o.SkirtRows-1, fs-2) // keep at least the feet row
	left, right := fs, -1
	for y := top; y <= hem; y++ {
		for x := 0; x < fs; x++ {
			if _, ok := rgbAt(pants, x, y); ok {
				left, right = min(left, x), max(right, x)
			}
		}
	}
	for y := top; y <= hem; y++ {
		flare := 0
		if y > top {
			flare = 1 // A-line: straight at the waist, 1px wider below
		}
		l, rr := max(left-flare, 0), min(right+flare, fs-1)
		for x := l; x <= rr; x++ {
			c := g[2]
			switch {
			case x == l || x == rr:
				c = d.outline
			case y == hem:
				c = t[1] // trim band at the hem
				if x == l+1 {
					c = t[2]
				} else if x == rr-1 {
					c = t[0]
				}
			case x == l+1:
				c = g[3] // lit from the top left
			case x == rr-1:
				c = g[1]
			case y == hem-1 && (x-l)%3 == 0:
				c = g[1] // a fold above the hem
			}
			setRGB(img, x, y, c)
		}
	}
	// Below the hem: trousers become bare legs; shoes and outline stay.
	for y := hem + 1; y < fs; y++ {
		for x := 0; x < fs; x++ {
			if c, ok := rgbAt(pants, x, y); ok {
				if i, cloth := d.pantsIdx[c]; cloth {
					c = d.skin[min(i, len(d.skin)-1)]
				}
				setRGB(img, x, y, c)
			}
		}
	}
	return img, nil
}
