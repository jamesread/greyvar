package rig

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Colour references let agents and people edit parts symbolically, so a new
// part can only ever contain rig colours:
//
//	skin.0 .. skin.N   key ramp entries (0 = darkest)
//	fixed.0 .. fixed.N fixed colours (outline etc.)
//	clear              transparent

// ColourRef resolves a symbolic colour. ok=false with err=nil means clear.
func (r *Rig) ColourRef(ref string) (RGB, bool, error) {
	if ref == "clear" {
		return RGB{}, false, nil
	}
	name, idxStr, found := strings.Cut(ref, ".")
	if !found {
		return RGB{}, false, fmt.Errorf("colour %q: want <slot>.<n>, fixed.<n> or clear", ref)
	}
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return RGB{}, false, fmt.Errorf("colour %q: %w", ref, err)
	}
	ramp := r.Spec.Slots[name]
	if name == "fixed" {
		ramp = r.Spec.Fixed
	}
	if ramp == nil {
		return RGB{}, false, fmt.Errorf("colour %q: unknown slot %q", ref, name)
	}
	if idx < 0 || idx >= len(ramp) {
		return RGB{}, false, fmt.Errorf("colour %q: index out of range 0..%d", ref, len(ramp)-1)
	}
	c, err := ParseHex(ramp[idx])
	return c, true, err
}

// refNames maps every rig colour to its symbolic name.
func (r *Rig) refNames() map[RGB]string {
	m := map[RGB]string{}
	for _, slot := range r.Spec.SlotOrder {
		for i, h := range r.Spec.Slots[slot] {
			if c, err := ParseHex(h); err == nil {
				m[c] = fmt.Sprintf("%s.%d", slot, i)
			}
		}
	}
	for i, h := range r.Spec.Fixed {
		if c, err := ParseHex(h); err == nil {
			m[c] = fmt.Sprintf("fixed.%d", i)
		}
	}
	return m
}

// ShowPart renders a part as a coordinate grid with a symbolic legend.
func (r *Rig) ShowPart(layer, variant, part string) (string, error) {
	img, err := ReadPNG(r.PartPath(layer, variant, part))
	if err != nil {
		return "", err
	}
	names := r.refNames()
	symbols := map[string]string{}
	var legend []string
	next := 'A'
	var b strings.Builder
	fmt.Fprintf(&b, "%s/%s/%s\n    ", layer, variant, part)
	for x := 0; x < img.Bounds().Dx(); x++ {
		fmt.Fprintf(&b, "%2d", x%10)
	}
	b.WriteString("\n")
	for y := 0; y < img.Bounds().Dy(); y++ {
		fmt.Fprintf(&b, "%2d  ", y)
		for x := 0; x < img.Bounds().Dx(); x++ {
			c, ok := rgbAt(img, x, y)
			if !ok {
				b.WriteString(" .")
				continue
			}
			name, known := names[c]
			if !known {
				name = c.Hex() + " (NOT A RIG COLOUR)"
			}
			sym, seen := symbols[name]
			if !seen {
				sym = string(next)
				next++
				symbols[name] = sym
				legend = append(legend, fmt.Sprintf("  %s = %s", sym, name))
			}
			b.WriteString(" " + sym)
		}
		b.WriteString("\n")
	}
	b.WriteString(strings.Join(legend, "\n"))
	b.WriteString("\n")
	return b.String(), nil
}

// PixelEdit sets one pixel (or a horizontal run X..X2) to a colour reference.
type PixelEdit struct {
	X, X2, Y int
	Ref      string
}

// ParsePixelEdit parses "x,y=ref" or "x1-x2,y=ref".
func ParsePixelEdit(s string) (PixelEdit, error) {
	pos, ref, ok := strings.Cut(s, "=")
	if !ok {
		return PixelEdit{}, fmt.Errorf("edit %q: want x,y=colour", s)
	}
	xs, ys, ok := strings.Cut(pos, ",")
	if !ok {
		return PixelEdit{}, fmt.Errorf("edit %q: want x,y=colour", s)
	}
	e := PixelEdit{Ref: strings.TrimSpace(ref)}
	var err error
	if e.Y, err = strconv.Atoi(strings.TrimSpace(ys)); err != nil {
		return e, fmt.Errorf("edit %q: %w", s, err)
	}
	x1, x2, isRange := strings.Cut(strings.TrimSpace(xs), "-")
	if e.X, err = strconv.Atoi(x1); err != nil {
		return e, fmt.Errorf("edit %q: %w", s, err)
	}
	e.X2 = e.X
	if isRange {
		if e.X2, err = strconv.Atoi(x2); err != nil {
			return e, fmt.Errorf("edit %q: %w", s, err)
		}
	}
	return e, nil
}

// DerivePart writes a new part: a copy of `from` (or blank when from is
// empty) with the edits applied. It refuses to overwrite unless force is set.
func (r *Rig) DerivePart(layer, variant, from, to string, edits []PixelEdit, force bool) (string, error) {
	fs := r.Spec.FrameSize
	img := image.NewNRGBA(image.Rect(0, 0, fs, fs))
	if from != "" {
		src, err := ReadPNG(r.PartPath(layer, variant, from))
		if err != nil {
			return "", err
		}
		img = src
	}
	for _, e := range edits {
		c, opaque, err := r.ColourRef(e.Ref)
		if err != nil {
			return "", err
		}
		for x := e.X; x <= e.X2; x++ {
			if x < 0 || e.Y < 0 || x >= fs || e.Y >= fs {
				return "", fmt.Errorf("edit (%d,%d) outside %dx%d part", x, e.Y, fs, fs)
			}
			if opaque {
				setRGB(img, x, e.Y, c)
			} else {
				img.SetNRGBA(x, e.Y, color.NRGBA{})
			}
		}
	}
	out := r.PartPath(layer, variant, to)
	if _, err := os.Stat(out); err == nil && !force {
		return "", fmt.Errorf("%s exists (use -force to overwrite)", out)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return "", err
	}
	return out, os.WriteFile(out, EncodePNG(img), 0o644)
}

// UnusedParts lists part files (any layer, any variant) that no pose
// references, sorted. They are safe to delete.
func (r *Rig) UnusedParts() ([]string, error) {
	used := map[string]bool{}
	for _, p := range r.Poses {
		for layer, part := range p.Parts {
			used[layer+"/"+part] = true
		}
	}
	files, err := filepath.Glob(filepath.Join(r.Dir, "parts", "*", "*", "*.png"))
	if err != nil {
		return nil, err
	}
	var unused []string
	for _, f := range files {
		rel, _ := filepath.Rel(filepath.Join(r.Dir, "parts"), f)
		segs := strings.Split(filepath.ToSlash(rel), "/")
		if len(segs) != 3 {
			continue
		}
		if !used[segs[0]+"/"+strings.TrimSuffix(segs[2], ".png")] {
			unused = append(unused, f)
		}
	}
	sort.Strings(unused)
	return unused, nil
}
