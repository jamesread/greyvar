package rig

import (
	"bytes"
	"fmt"
	"os"
	"sort"
)

// Validate returns human-readable problems; empty means the rig is valid.
func (r *Rig) Validate() []string {
	var errs []string
	add := func(format string, args ...any) { errs = append(errs, fmt.Sprintf(format, args...)) }
	s := r.Spec
	fs := s.FrameSize

	if s.Name == "" {
		add("rig.yml: name is required")
	}
	if s.Outputs.SheetsDir == "" || s.Outputs.Composite == "" || s.Outputs.WebRoot == "" {
		add("rig.yml: outputs.sheetsDir, outputs.composite and outputs.webRoot are required")
	}

	if hb := s.HitboxOrDefault(); hb[2] <= 0 || hb[3] <= 0 || hb[0] < 0 || hb[1] < 0 || hb[0]+hb[2] > CellSize || hb[1]+hb[3] > CellSize {
		add("rig.yml: hitbox %v must have positive size and fit in the %dx%d cell", hb, CellSize, CellSize)
	}
	if s.Size != "child" && s.Size != "adult" {
		add("rig.yml: size %q must be child or adult", s.Size)
	}
	if a := s.AnchorOrDefault(); a[0] < 0 || a[1] < 0 || a[0] > fs || a[1] > fs {
		add("rig.yml: anchor %v must be inside the %dx%d frame", a, fs, fs)
	}

	// Colours.
	allowed := map[RGB]string{}
	addColour := func(hex, owner string) {
		c, err := ParseHex(hex)
		if err != nil {
			add("rig.yml %s: %v", owner, err)
			return
		}
		if prev, dup := allowed[c]; dup {
			add("rig.yml: colour %s used by both %s and %s; key colours must be unique", hex, prev, owner)
		}
		allowed[c] = owner
	}
	if len(s.SlotOrder) != len(s.Slots) {
		add("rig.yml: slotOrder must list every slot exactly once")
	}
	for _, slot := range s.SlotOrder {
		ramp, ok := s.Slots[slot]
		if !ok {
			add("rig.yml: slotOrder names unknown slot %q", slot)
			continue
		}
		for i, hex := range ramp {
			addColour(hex, fmt.Sprintf("slots.%s[%d]", slot, i))
		}
		pals := s.Palettes[slot]
		if len(pals) == 0 {
			add("rig.yml: slot %s has no palettes", slot)
		}
		for name, p := range pals {
			if len(p) != len(ramp) {
				add("rig.yml: palette %s.%s has %d colours, slot has %d", slot, name, len(p), len(ramp))
			}
			for _, hex := range p {
				if _, err := ParseHex(hex); err != nil {
					add("rig.yml: palette %s.%s: %v", slot, name, err)
				}
			}
		}
		for _, h := range s.Hidden[slot] {
			if _, ok := pals[h]; !ok {
				add("rig.yml: hidden palette %s.%s is not a palette", slot, h)
			}
			if s.Defaults.Palettes[slot] == h {
				add("rig.yml: defaults.palettes.%s cannot be hidden palette %q", slot, h)
			}
		}
		if d := s.Defaults.Palettes[slot]; d == "" {
			add("rig.yml: defaults.palettes.%s missing", slot)
		} else if _, ok := pals[d]; !ok {
			add("rig.yml: defaults.palettes.%s = %q is not a palette", slot, d)
		}
	}
	for i, hex := range s.Fixed {
		addColour(hex, fmt.Sprintf("fixed[%d]", i))
	}

	// Layers.
	layerNames := map[string]bool{}
	for _, l := range s.Layers {
		if layerNames[l.Name] {
			add("rig.yml: duplicate layer %q", l.Name)
		}
		layerNames[l.Name] = true
		if d, ok := s.Defaults.Parts[l.Name]; ok && !contains(LayerVariants(l), d) {
			add("rig.yml: defaults.parts.%s = %q is not a variant", l.Name, d)
		}
	}

	// Poses and their parts.
	poseNames := map[string]bool{}
	pc := partCache{}
	for _, p := range r.Poses {
		if poseNames[p.Name] {
			add("poses.yml: duplicate pose %q", p.Name)
		}
		poseNames[p.Name] = true
		if p.ItemZ != "" && p.ItemZ != "front" && p.ItemZ != "behind" {
			add("pose %s: itemZ must be front or behind", p.Name)
		}
		if p.Hand != nil && (p.Hand[0] < 0 || p.Hand[1] < 0 || p.Hand[0] >= fs || p.Hand[1] >= fs) {
			add("pose %s: hand %v outside %dx%d frame", p.Name, *p.Hand, fs, fs)
		}
		for layer := range p.Offsets {
			if !layerNames[layer] {
				add("pose %s: offset for unknown layer %q", p.Name, layer)
			}
		}
		for _, layer := range sortedKeys(p.Parts) {
			part := p.Parts[layer]
			l, ok := r.Layer(layer)
			if !ok {
				add("pose %s: unknown layer %q", p.Name, layer)
				continue
			}
			off := p.Offsets[layer]
			for _, v := range LayerVariants(l) {
				img, err := pc.get(r, layer, v, part)
				if err != nil {
					add("pose %s: layer %s variant %s: missing part %s (%v)", p.Name, layer, v, part, err)
					continue
				}
				b := img.Bounds()
				if b.Dx() != fs || b.Dy() != fs {
					add("part %s/%s/%s: is %dx%d, want %dx%d", layer, v, part, b.Dx(), b.Dy(), fs, fs)
					continue
				}
				for y := 0; y < fs; y++ {
					for x := 0; x < fs; x++ {
						c := img.NRGBAAt(x, y)
						if c.A == 0 {
							continue
						}
						if c.A != 255 {
							add("part %s/%s/%s: pixel (%d,%d) is semi-transparent (a=%d)", layer, v, part, x, y, c.A)
							continue
						}
						if _, ok := allowed[RGB{c.R, c.G, c.B}]; !ok {
							add("part %s/%s/%s: pixel (%d,%d) colour %s is not a rig colour", layer, v, part, x, y, RGB{c.R, c.G, c.B}.Hex())
						}
						tx, ty := x+off[0], y+off[1]
						if tx < 0 || ty < 0 || tx >= fs || ty >= fs {
							add("pose %s: layer %s offset %v pushes pixel (%d,%d) out of frame", p.Name, layer, off, x, y)
						}
					}
				}
			}
		}
	}

	// Clips.
	for _, name := range sortedKeys(r.Clips.Clips) {
		steps := r.Clips.Clips[name]
		if len(steps) == 0 {
			add("clip %s: no steps", name)
		}
		for i, st := range steps {
			if !poseNames[st.Pose] {
				add("clip %s step %d: unknown pose %q", name, i, st.Pose)
			}
			if st.Ms <= 0 {
				add("clip %s step %d: ms must be > 0", name, i)
			}
		}
	}
	for _, state := range sortedKeys(r.Clips.States) {
		if _, ok := r.Clips.Clips[r.Clips.States[state]]; !ok {
			add("state %s: unknown clip %q", state, r.Clips.States[state])
		}
	}
	if _, ok := r.Clips.States[r.Clips.InitialState]; !ok {
		add("clips.yml: initialState %q is not in states", r.Clips.InitialState)
	}
	idle := r.Clips.Idle
	if idle.Base != "" {
		if _, ok := r.Clips.Clips[idle.Base]; !ok {
			add("idle.base: unknown clip %q", idle.Base)
		}
		for _, v := range idle.Variants {
			if _, ok := r.Clips.Clips[v.Clip]; !ok {
				add("idle variant: unknown clip %q", v.Clip)
			}
			if v.Weight <= 0 {
				add("idle variant %s: weight must be > 0", v.Clip)
			}
		}
		if len(idle.Variants) > 0 && (idle.IntervalMs[0] <= 0 || idle.IntervalMs[1] < idle.IntervalMs[0]) {
			add("idle.intervalMs must be [min, max] with 0 < min <= max")
		}
	}
	return errs
}

// Stale compares generated outputs with what is on disk and returns the
// paths that differ or are missing.
func Stale(outs []Output) []string {
	var stale []string
	for _, o := range outs {
		disk, err := os.ReadFile(o.Path)
		if err != nil || !bytes.Equal(disk, o.Data) {
			stale = append(stale, o.Path)
		}
	}
	return stale
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
