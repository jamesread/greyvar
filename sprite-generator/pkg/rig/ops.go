package rig

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
)

// Variant ops create a whole new layer variant from an existing one in one
// deterministic pass. The output parts are ordinary source PNGs: review them
// with `spritegen preview` and hand-fix any part with `part derive -force`.

// LongHairOptions configures the long-hair op.
type LongHairOptions struct {
	Layer, From, To string
	Rows            int // how far to extend below the existing hairline
	Force           bool
}

// DeriveLongHair writes layer/To/<part>.png for every part used by a pose.
// Front/side views (pose has a head part): the outermost two hair columns on
// each side grow down. Back views (no head part): every hair column grows.
func (r *Rig) DeriveLongHair(o LongHairOptions) ([]string, error) {
	if o.Rows <= 0 {
		o.Rows = 3
	}
	outline, _, err := r.ColourRef("fixed.0")
	if err != nil {
		return nil, err
	}
	slot := o.Layer // hair layer uses the hair slot
	keys := r.Spec.Slots[slot]
	if len(keys) == 0 {
		return nil, fmt.Errorf("layer %s has no colour slot named %s", o.Layer, slot)
	}
	mid, _ := ParseHex(keys[len(keys)/2])
	dark, _ := ParseHex(keys[0])
	isHair := map[RGB]bool{}
	for _, k := range keys {
		c, _ := ParseHex(k)
		isHair[c] = true
	}

	// part name -> whether any pose using it is a back view
	back := map[string]bool{}
	for _, p := range r.Poses {
		part, ok := p.Parts[o.Layer]
		if !ok {
			continue
		}
		if _, hasHead := p.Parts["head"]; !hasHead {
			back[part] = true
		} else if _, seen := back[part]; !seen {
			back[part] = false
		}
	}
	names := make([]string, 0, len(back))
	for n := range back {
		names = append(names, n)
	}
	sort.Strings(names)

	fs := r.Spec.FrameSize
	var written []string
	for _, name := range names {
		src, err := ReadPNG(r.PartPath(o.Layer, o.From, name))
		if err != nil {
			return nil, err
		}
		img := image.NewNRGBA(src.Bounds())
		copy(img.Pix, src.Pix)

		// Hair mass: columns with hair colour in the top 6 rows (ignores
		// stray import pixels lower down).
		bottom := map[int]int{}
		for x := 0; x < fs; x++ {
			for y := 0; y < 6; y++ {
				if c, ok := rgbAt(src, x, y); ok && isHair[c] {
					bottom[x] = y
				}
			}
		}
		if len(bottom) == 0 {
			continue
		}
		cols := make([]int, 0, len(bottom))
		for x := range bottom {
			cols = append(cols, x)
		}
		sort.Ints(cols)
		grow := cols
		if !back[name] && len(cols) > 4 {
			grow = []int{cols[0], cols[1], cols[len(cols)-2], cols[len(cols)-1]}
		}
		growSet := map[int]bool{}
		for _, x := range grow {
			growSet[x] = true
		}
		for _, x := range grow {
			end := min(bottom[x]+o.Rows, fs-2)
			for y := bottom[x] + 1; y <= end; y++ {
				c := mid
				if y == end || x == cols[0] || x == cols[len(cols)-1] {
					c = dark
				}
				setRGB(img, x, y, c)
			}
			// Outline below and on exposed sides.
			setRGB(img, x, end+1, outline)
			for _, side := range []int{x - 1, x + 1} {
				if side < 0 || side >= fs || growSet[side] {
					continue
				}
				for y := bottom[x] + 1; y <= end; y++ {
					// Only outer edges: an inner outline would cover the face.
					if side < cols[0] || side > cols[len(cols)-1] {
						setRGB(img, side, y, outline)
					}
				}
			}
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
