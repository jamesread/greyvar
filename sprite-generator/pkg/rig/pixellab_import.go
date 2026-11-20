package rig

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// PixelLab import: a character made with PixelLab (create_character in pro
// mode with its palette locked to the rig's key colours, then template
// animations) is split into rig layers by colour. Source frames are named
//
//	<anim>_<dir>_<i>.png   anim: idle | breathe | pockets | walk; dir: south | east | north | west
//
// idle is the still standing pose (idle_<dir>_0; bump, use and the `stand`
// state show it); breathe, when present, is the looping idle animation;
// pockets, when present, is a second breathing stance (hands in pockets:
// clips pockets, pocketsUp, pocketsLeft, pocketsRight and state `pockets`).
//
// and are pivot-centred on a square canvas. Every frame is placed on the
// rig's frame (feet on the anchor's baseline), each pixel snapped to the
// nearest key colour, and pixels sorted into layers by slot:
//
//	skin and eyes -> body, hair -> hair, shirt -> shirt, pants -> pants,
//	garment and trim (when the rig has them) -> outfit
//
// Outline pixels go to the layer of the pixels they border; the second
// fixed colour (eyes) is kept as eyes only inside the face, elsewhere it is
// outline (PixelLab uses it as a soft inner outline too).
//
// rig.yml (slots, palettes, layers, outputs) is written by hand first; the
// import writes the parts, poses.yml and clips.yml.

// PixelLabImportOptions configures the import.
type PixelLabImportOptions struct {
	Source  string            // directory of <anim>_<dir>_<i>.png
	Variant map[string]string // layer -> variant the parts are written to (default "default")
	Layers  []string          // layers to write (default all four)
	WalkMs  int               // ms per walk frame (default 100)
	IdleMs  int               // ms per idle frame (default 250)
	// BreatheMs is ms per breathing frame (default 600: four frames make a
	// calm 2.4 s breath).
	BreatheMs int
	Force   bool
	// Poses writes poses.yml and clips.yml too (the base character only;
	// variants such as hairstyles reuse the base's poses).
	Poses bool
	// NoStabilize keeps PixelLab's frames where they are. By default every
	// frame of an animation is shifted sideways so its head lines up with
	// the first frame's: PixelLab redraws each frame and the figure drifts.
	NoStabilize bool
}

// headCentre is the horizontal centre of the top rows (the head) of a frame,
// in source pixels, or -1 for an empty frame.
func headCentre(img *image.NRGBA) int {
	b := opaqueBounds(img)
	if b.Empty() {
		return -1
	}
	rows := max(4, b.Dy()/3)
	sum, n := 0, 0
	for y := b.Min.Y; y < b.Min.Y+rows && y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.NRGBAAt(x, y).A > 0 {
				sum += x
				n++
			}
		}
	}
	if n == 0 {
		return -1
	}
	return (sum + n/2) / n
}

var pixellabDirs = []struct{ src, facing string }{
	{"south", "Down"}, {"east", "Right"}, {"north", "Up"}, {"west", "Left"},
}

type pxClass int

const (
	pxNone pxClass = iota
	pxOutline
	pxEye
	pxSkin
	pxHair
	pxShirt
	pxPants
	pxGarment
)

var pxClassLayer = map[pxClass]string{pxSkin: "body", pxEye: "body", pxHair: "hair", pxShirt: "shirt", pxPants: "pants", pxGarment: "outfit"}

// ImportPixelLab writes the parts (and, with Poses, poses.yml/clips.yml).
func (r *Rig) ImportPixelLab(o PixelLabImportOptions) ([]string, error) {
	if o.WalkMs == 0 {
		o.WalkMs = 100
	}
	if o.IdleMs == 0 {
		o.IdleMs = 250
	}
	if o.BreatheMs == 0 {
		o.BreatheMs = 600
	}
	if len(o.Layers) == 0 {
		o.Layers = []string{"pants", "body", "shirt", "hair"}
	}
	variant := func(layer string) string {
		if v, ok := o.Variant[layer]; ok {
			return v
		}
		return "default"
	}

	// Key colour -> class.
	type key struct {
		c   RGB
		cls pxClass
	}
	var keys []key
	addSlot := func(slot string, cls pxClass) error {
		ks, ok := r.Spec.Slots[slot]
		if !ok {
			return fmt.Errorf("rig.yml has no %s slot", slot)
		}
		for _, h := range ks {
			c, err := ParseHex(h)
			if err != nil {
				return err
			}
			keys = append(keys, key{c, cls})
		}
		return nil
	}
	for slot, cls := range map[string]pxClass{"skin": pxSkin, "hair": pxHair, "shirt": pxShirt, "pants": pxPants} {
		if err := addSlot(slot, cls); err != nil {
			return nil, err
		}
	}
	// Outfit colours (a dress: garment and trim keys) -> the outfit layer.
	for _, slot := range []string{"garment", "trim"} {
		if _, ok := r.Spec.Slots[slot]; ok {
			if err := addSlot(slot, pxGarment); err != nil {
				return nil, err
			}
		}
	}
	if len(r.Spec.Fixed) < 2 {
		return nil, fmt.Errorf("rig.yml needs two fixed colours (outline, eyes)")
	}
	outline, _ := ParseHex(r.Spec.Fixed[0])
	eye, _ := ParseHex(r.Spec.Fixed[1])
	keys = append(keys, key{outline, pxOutline}, key{eye, pxEye})
	sort.Slice(keys, func(i, j int) bool { return keys[i].c.Hex() < keys[j].c.Hex() })
	snap := func(c RGB) key {
		best, bd := keys[0], 1<<30
		for _, k := range keys {
			dr, dg, db := int(c.R)-int(k.c.R), int(c.G)-int(k.c.G), int(c.B)-int(k.c.B)
			if d := dr*dr + dg*dg + db*db; d < bd {
				best, bd = k, d
			}
		}
		return best
	}

	// Frames and their names, in pose order.
	type frame struct {
		name, anim, facing string
		img                *image.NRGBA
	}
	var frames []frame
	for _, anim := range []string{"idle", "breathe", "pockets", "walk"} {
		for _, d := range pixellabDirs {
			for i := 0; ; i++ {
				p := filepath.Join(o.Source, fmt.Sprintf("%s_%s_%d.png", anim, d.src, i))
				if _, err := os.Stat(p); err != nil {
					break
				}
				img, err := ReadPNG(p)
				if err != nil {
					return nil, err
				}
				frames = append(frames, frame{fmt.Sprintf("%s_%s_%d", anim, d.src, i), anim, d.facing, img})
			}
		}
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("no <anim>_<dir>_<i>.png frames in %s", o.Source)
	}

	// Place every frame centred on the anchor with its feet on the anchor
	// line. PixelLab's rotations and animations come on different canvases
	// (e.g. 28px and 48px), so the feet line is found per canvas size.
	fs := r.Spec.FrameSize
	anchor := r.Spec.AnchorOrDefault()
	feet := map[image.Point]int{}
	for _, f := range frames {
		sz := f.img.Bounds().Size()
		if b := opaqueBounds(f.img); b.Max.Y > feet[sz] {
			feet[sz] = b.Max.Y
		}
	}
	place := func(img *image.NRGBA) (int, int) {
		sz := img.Bounds().Size()
		return anchor[0] - sz.X/2, anchor[1] - feet[sz]
	}

	// Stabilise: line each animation's frames up on its first frame's head.
	shiftX := map[string]int{}
	if !o.NoStabilize {
		ref := map[string]int{}
		for _, f := range frames {
			key := f.anim + "_" + f.facing
			c := headCentre(f.img)
			if c < 0 {
				continue
			}
			if r, ok := ref[key]; ok {
				shiftX[f.name] = r - c
			} else {
				ref[key] = c
			}
		}
	}

	var written []string
	for _, f := range frames {
		W, H := fs, fs
		cls := make([][]pxClass, H)
		col := make([][]RGB, H)
		for y := range cls {
			cls[y], col[y] = make([]pxClass, W), make([]RGB, W)
		}
		src := f.img.Bounds()
		dx, dy := place(f.img)
		for y := 0; y < src.Dy(); y++ {
			for x := 0; x < src.Dx(); x++ {
				c, ok := rgbAt(f.img, x, y)
				if !ok {
					continue
				}
				X, Y := x+dx+shiftX[f.name], y+dy
				if X < 0 || Y < 0 || X >= W || Y >= H {
					return nil, fmt.Errorf("%s: pixel (%d,%d) falls outside the %dpx frame; raise frameSize or move the anchor", f.name, x, y, fs)
				}
				k := snap(c)
				cls[Y][X], col[Y][X] = k.cls, k.c
			}
		}
		// Eyes only inside the face: an eye pixel with skin on both sides
		// (horizontally) or above and below; else it is outline.
		isSkin := func(x, y int) bool { return x >= 0 && y >= 0 && x < W && y < H && cls[y][x] == pxSkin }
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if cls[y][x] == pxEye && !((isSkin(x-1, y) || isSkin(x-2, y)) && (isSkin(x+1, y) || isSkin(x+2, y))) {
					cls[y][x], col[y][x] = pxOutline, outline
				}
			}
		}
		// PixelLab often draws eyes in a dark hair shade: a run of one or two
		// hair pixels with skin directly left and right of it is an eye, so
		// it must not take the hair colour (blonde eyes) or go with the
		// hairstyle (a bald head would have holes for eyes).
		isHair := func(x, y int) bool { return x >= 0 && y >= 0 && x < W && y < H && cls[y][x] == pxHair }
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if !isHair(x, y) || !isSkin(x-1, y) {
					continue
				}
				switch {
				case isSkin(x+1, y):
					cls[y][x], col[y][x] = pxEye, eye
				case isHair(x+1, y) && isSkin(x+2, y):
					cls[y][x], col[y][x] = pxEye, eye
					cls[y][x+1], col[y][x+1] = pxEye, eye
				}
			}
		}
		// Outline -> the layer most of its non-outline neighbours belong to.
		owner := make([][]string, H)
		for y := range owner {
			owner[y] = make([]string, W)
			for x := 0; x < W; x++ {
				owner[y][x] = pxClassLayer[cls[y][x]]
			}
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if cls[y][x] != pxOutline {
					continue
				}
				votes := map[string]int{}
				for ry := -1; ry <= 1; ry++ {
					for rx := -1; rx <= 1; rx++ {
						X, Y := x+rx, y+ry
						if X >= 0 && Y >= 0 && X < W && Y < H {
							if l := pxClassLayer[cls[Y][X]]; l != "" {
								votes[l]++
							}
						}
					}
				}
				best, bv := "body", 0
				for _, l := range []string{"hair", "outfit", "shirt", "pants", "body"} { // tie order
					if votes[l] > bv {
						best, bv = l, votes[l]
					}
				}
				owner[y][x] = best
			}
		}
		for _, layer := range o.Layers {
			img := image.NewNRGBA(image.Rect(0, 0, W, H))
			for y := 0; y < H; y++ {
				for x := 0; x < W; x++ {
					if cls[y][x] != pxNone && owner[y][x] == layer {
						setRGB(img, x, y, col[y][x])
					}
				}
			}
			out := r.PartPath(layer, variant(layer), f.name)
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
	}

	if !o.Poses {
		return written, nil
	}
	// poses.yml: one pose per frame (every rig layer uses the frame's part,
	// the outfit layer its own), plus 1px bump leans; clips.yml from them.
	var poses []Pose
	byAnim := map[string][]string{} // "<anim>_<facing>" -> pose names
	for _, f := range frames {
		parts := map[string]string{}
		for _, l := range r.Spec.Layers {
			parts[l.Name] = f.name
		}
		poses = append(poses, Pose{Name: f.name, Parts: parts})
		byAnim[f.anim+"_"+f.facing] = append(byAnim[f.anim+"_"+f.facing], f.name)
	}
	lean := map[string][2]int{"Down": {0, 1}, "Up": {0, -1}, "Left": {-1, 0}, "Right": {1, 0}}
	for _, d := range pixellabDirs {
		if len(byAnim["idle_"+d.facing])+len(byAnim["breathe_"+d.facing])+len(byAnim["walk_"+d.facing]) == 0 {
			return nil, fmt.Errorf("no idle or walk frames facing %s", d.src)
		}
		base := poses[indexOfPose(poses, stand(byAnim, d.facing))]
		off := map[string][2]int{}
		for _, l := range r.Spec.Layers {
			off[l.Name] = lean[d.facing]
		}
		poses = append(poses, Pose{Name: "bump" + d.facing, Parts: base.Parts, Offsets: off})
	}
	clips := ClipsSpec{Clips: map[string][]Step{}, States: map[string]string{}, InitialState: "idle"}
	steps := func(names []string, ms int) []Step {
		var out []Step
		for _, n := range names {
			out = append(out, Step{Pose: n, Ms: ms})
		}
		return out
	}
	side := func(f string) string { // diagonals use the side view, like the child rig
		if strings.Contains(f, "Left") {
			return "Left"
		}
		if strings.Contains(f, "Right") {
			return "Right"
		}
		return f
	}
	for _, f := range []string{"Down", "Up", "Left", "Right", "DownLeft", "DownRight", "UpLeft", "UpRight"} {
		face := side(f)
		walk := byAnim["walk_"+face]
		clips.Clips["walk"+f] = steps(walk, o.WalkMs)
		clips.Clips["bump"+f] = []Step{{Pose: "bump" + face, Ms: 120}, {Pose: stand(byAnim, face), Ms: 120}}
		clips.States["walk"+f] = "walk" + f
	}
	for _, f := range []string{"Down", "Up", "Left", "Right"} {
		idle, ms := byAnim["breathe_"+f], o.BreatheMs
		if len(idle) == 0 {
			idle, ms = byAnim["idle_"+f], o.IdleMs
		}
		if len(idle) == 0 {
			idle = []string{stand(byAnim, f)}
		}
		name := "idle" + f
		if f == "Down" {
			name = "idle"
		}
		clips.Clips[name] = steps(idle, ms)
		clips.Clips["use"+f] = []Step{{Pose: "bump" + f, Ms: 150}, {Pose: stand(byAnim, f), Ms: 150}}
	}
	clips.States["idle"] = "idle"
	// stand: the still standing pose, for figures that must not breathe
	// (mannequins).
	clips.Clips["stand"] = []Step{{Pose: stand(byAnim, "Down"), Ms: 1000}}
	clips.States["stand"] = "stand"
	if len(byAnim["pockets_Down"]) > 0 {
		for _, f := range []string{"Down", "Up", "Left", "Right"} {
			name := "pockets" + f
			if f == "Down" {
				name = "pockets"
			}
			frames := byAnim["pockets_"+f]
			if len(frames) == 0 {
				frames = []string{stand(byAnim, f)}
			}
			clips.Clips[name] = steps(frames, o.BreatheMs)
		}
		clips.States["pockets"] = "pockets"
	}
	clips.Idle = IdleSpec{Base: "idle", IntervalMs: [2]int{4000, 9000}}

	if err := writeYAMLFile(filepath.Join(r.Dir, "poses.yml"), "# Generated by `spritegen import-pixellab`; re-run it after changing the source frames.\n", posesFile{Poses: poses}); err != nil {
		return nil, err
	}
	if err := writeYAMLFile(filepath.Join(r.Dir, "clips.yml"), "# Generated by `spritegen import-pixellab`; re-run it after changing the source frames.\n", clips); err != nil {
		return nil, err
	}
	return written, nil
}

func stand(byAnim map[string][]string, facing string) string {
	if s := byAnim["idle_"+facing]; len(s) > 0 {
		return s[0]
	}
	return byAnim["walk_"+facing][0]
}

func indexOfPose(ps []Pose, name string) int {
	for i, p := range ps {
		if p.Name == name {
			return i
		}
	}
	return -1
}

func opaqueBounds(img *image.NRGBA) image.Rectangle {
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, -1, -1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.NRGBAAt(x, y).A > 0 {
				minX, minY, maxX, maxY = min(minX, x), min(minY, y), max(maxX, x), max(maxY, y)
			}
		}
	}
	if maxX < 0 {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

func writeYAMLFile(path, header string, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append([]byte(header), data...), 0o644)
}
