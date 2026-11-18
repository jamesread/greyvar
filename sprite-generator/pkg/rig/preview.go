package rig

import (
	"fmt"
	"image"
	"image/color"
)

// PreviewOptions selects what to render.
type PreviewOptions struct {
	Clip     string            // render this clip's steps in order; empty = every pose
	Parts    map[string]string // layer -> variant overrides
	Palettes map[string]string // slot -> palette overrides
	Scale    int
}

// Preview renders a strip of frames, upscaled, on a neutral background, for
// visual review. Frames are separated by a 1-frame-pixel gutter.
func (r *Rig) Preview(o PreviewOptions) (*image.NRGBA, error) {
	parts := map[string]string{}
	for k, v := range r.Spec.Defaults.Parts {
		parts[k] = v
	}
	for k, v := range o.Parts {
		parts[k] = v
	}
	pals := map[string]string{}
	for k, v := range r.Spec.Defaults.Palettes {
		pals[k] = v
	}
	for k, v := range o.Palettes {
		pals[k] = v
	}
	sheet, err := r.Compose(parts, pals)
	if err != nil {
		return nil, err
	}

	var frames []int
	if o.Clip != "" {
		steps, ok := r.Clips.Clips[o.Clip]
		if !ok {
			return nil, fmt.Errorf("unknown clip %q", o.Clip)
		}
		idx := r.PoseIndex()
		for _, s := range steps {
			frames = append(frames, idx[s.Pose])
		}
	} else {
		for i := range r.Poses {
			frames = append(frames, i)
		}
	}
	scale := o.Scale
	if scale <= 0 {
		scale = 8
	}
	fs, cols := r.Spec.FrameSize, r.Spec.Columns
	perRow := cols
	if o.Clip != "" {
		perRow = len(frames)
	}
	rows := (len(frames) + perRow - 1) / perRow
	cell := fs + 1
	out := image.NewNRGBA(image.Rect(0, 0, (perRow*cell+1)*scale, (rows*cell+1)*scale))
	bg := [2]color.NRGBA{{0xd8, 0xd8, 0xd8, 255}, {0xc8, 0xc8, 0xc8, 255}}
	gutter := color.NRGBA{0x80, 0x80, 0x90, 255}
	for y := 0; y < out.Bounds().Dy(); y++ {
		for x := 0; x < out.Bounds().Dx(); x++ {
			out.SetNRGBA(x, y, gutter)
		}
	}
	for n, f := range frames {
		ox, oy := (n%perRow)*cell+1, (n/perRow)*cell+1
		sx, sy := (f%cols)*fs, (f/cols)*fs
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				c := sheet.NRGBAAt(sx+x, sy+y)
				if c.A == 0 {
					c = bg[(x+y)%2]
				}
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						out.SetNRGBA((ox+x)*scale+dx, (oy+y)*scale+dy, c)
					}
				}
			}
		}
	}
	return out, nil
}
