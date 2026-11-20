package rig

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
// Long hair must never cover the face:
//   - back views (no head part): every hair column grows down;
//   - side views (pose name has Left/Right): only the back of the head grows,
//     the two hair columns away from the face;
//   - front views: two-column locks hang just outside the face on each side,
//     placed from the head part's skin pixels row by row.
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
	isSkin := map[RGB]bool{}
	for _, k := range r.Spec.Slots["skin"] {
		c, _ := ParseHex(k)
		isSkin[c] = true
	}

	// part name -> its view and the head part drawn with it
	type use struct{ view, head string }
	uses := map[string]use{}
	for _, p := range r.Poses {
		part, ok := p.Parts[o.Layer]
		if !ok {
			continue
		}
		if _, seen := uses[part]; seen {
			continue
		}
		head, hasHead := p.Parts["head"]
		u := use{view: "front", head: head}
		switch {
		case !hasHead:
			u.view = "back"
		case strings.Contains(strings.ToLower(p.Name), "left"):
			u.view = "left" // faces left: the back of the head is on the right
		case strings.Contains(strings.ToLower(p.Name), "right"):
			u.view = "right"
		}
		uses[part] = u
	}
	names := make([]string, 0, len(uses))
	for n := range uses {
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
		lowest := 0
		for _, y := range bottom {
			lowest = max(lowest, y)
		}

		// grow paints one lock: columns xs (inner to outer) from row y0 to end,
		// outlined below and on its outer side.
		grow := func(xs []int, outerSide int, y0, end int) {
			for i, x := range xs {
				if x < 0 || x >= fs {
					continue
				}
				for y := y0; y <= end; y++ {
					c := mid
					if y == end || i == len(xs)-1 {
						c = dark
					}
					setRGB(img, x, y, c)
				}
				setRGB(img, x, end+1, outline)
			}
			if outerSide >= 0 && outerSide < fs {
				for y := y0; y <= end+1; y++ {
					if _, opaque := rgbAt(img, outerSide, y); !opaque {
						setRGB(img, outerSide, y, outline)
					}
				}
			}
		}
		u := uses[name]
		end := min(lowest+o.Rows, fs-2)
		switch {
		case u.view == "back" || len(cols) <= 4:
			// Every column grows; dark on the last row and the outer columns.
			first, last := cols[0], cols[len(cols)-1]
			for _, x := range cols {
				colEnd := min(bottom[x]+o.Rows, fs-2)
				for y := bottom[x] + 1; y <= colEnd; y++ {
					c := mid
					if y == colEnd || x == first || x == last {
						c = dark
					}
					setRGB(img, x, y, c)
				}
				setRGB(img, x, colEnd+1, outline)
				for _, side := range []int{first - 1, last + 1} {
					if (x == first && side == first-1) || (x == last && side == last+1) {
						for y := bottom[x] + 1; y <= colEnd; y++ {
							if side >= 0 && side < fs {
								setRGB(img, side, y, outline)
							}
						}
					}
				}
			}
		case u.view == "left":
			n := len(cols)
			grow([]int{cols[n-2], cols[n-1]}, cols[n-1]+1, lowest+1, end)
		case u.view == "right":
			grow([]int{cols[1], cols[0]}, cols[0]-1, lowest+1, end)
		default: // front: frame the face from outside its skin
			head, err := ReadPNG(r.PartPath("head", "default", u.head))
			if err != nil {
				return nil, err
			}
			faceL, faceR := fs, -1
			for y := lowest + 1; y <= end; y++ {
				for x := 0; x < fs; x++ {
					if c, ok := rgbAt(head, x, y); ok && isSkin[c] {
						faceL, faceR = min(faceL, x), max(faceR, x)
					}
				}
			}
			if faceR < 0 { // no face found: fall back to the hair edges
				faceL, faceR = cols[0]+1, cols[len(cols)-1]-1
			}
			grow([]int{faceL - 1, faceL - 2}, faceL-3, lowest+1, end)
			grow([]int{faceR + 1, faceR + 2}, faceR+3, lowest+1, end)
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

// SkirtOptions configures the skirt op.
type SkirtOptions struct {
	Layer, From, To string
	Rows            int // skirt height from the top of the legs (default 2)
	Force           bool
}

// DeriveSkirt writes layer/To/<part>.png for every part used by a pose: the
// top Rows of each trousers part become a skirt (the last row flared 1px each
// side, a darker hem, outlined) and the trouser colour left below the hem
// becomes bare skin (same shade index). Outline pixels (shoes, leg edges)
// stay, so walk cycles still show the feet moving.
func (r *Rig) DeriveSkirt(o SkirtOptions) ([]string, error) {
	if o.Rows <= 0 {
		o.Rows = 2
	}
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
	cloth, skin := ramp(o.Layer), ramp("skin")
	if len(cloth) < 2 || len(skin) < 2 {
		return nil, fmt.Errorf("skirt needs the %s and skin colour slots", o.Layer)
	}
	clothIndex := map[RGB]int{}
	for i, c := range cloth {
		clothIndex[c] = i
	}
	parts := map[string]bool{}
	for _, p := range r.Poses {
		if part, ok := p.Parts[o.Layer]; ok {
			parts[part] = true
		}
	}
	names := make([]string, 0, len(parts))
	for n := range parts {
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
		top, left, right := -1, fs, -1
		for y := 0; y < fs && top < 0; y++ {
			for x := 0; x < fs; x++ {
				if _, ok := rgbAt(src, x, y); ok {
					top = y
					break
				}
			}
		}
		if top < 0 {
			continue
		}
		hem := min(top+o.Rows-1, fs-2)
		for y := top; y <= hem; y++ {
			for x := 0; x < fs; x++ {
				if _, ok := rgbAt(src, x, y); ok {
					left, right = min(left, x), max(right, x)
				}
			}
		}
		// Skirt rows: straight at the hips, flared on the hem row.
		for y := top; y <= hem; y++ {
			l, rr := left, right
			if y == hem && y > top {
				l, rr = max(left-1, 0), min(right+1, fs-1)
			}
			for x := l; x <= rr; x++ {
				c := cloth[len(cloth)/2]
				if y == hem && y > top {
					c = cloth[0] // darker hem
				}
				if x == l || x == rr {
					c = outline
				}
				setRGB(img, x, y, c)
			}
		}
		// Below the hem: trousers become bare legs.
		for y := hem + 1; y < fs; y++ {
			for x := 0; x < fs; x++ {
				if c, ok := rgbAt(src, x, y); ok {
					if i, isCloth := clothIndex[c]; isCloth {
						setRGB(img, x, y, skin[min(i, len(skin)-1)])
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
