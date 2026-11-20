package rig

import (
	"path/filepath"
	"testing"
)

// The adult rig (res/sprites/adult) as built from its PixelLab sources by
// res/sprites/adult/source/assemble.py.

func loadAdult(t *testing.T) *Rig {
	t.Helper()
	r, err := Load(filepath.Join("..", "..", "..", "res", "sprites", "adult"))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func layerVariants(r *Rig, layer string) []string {
	for _, l := range r.Spec.Layers {
		if l.Name == layer {
			return l.Variants
		}
	}
	return nil
}

// No hairstyle or beard covers the eyes, in any pose. An eye is an
// eye-coloured body pixel that the short hair (the look the base was drawn
// with) leaves visible; the bald head's ear line uses the same colour.
func TestAdultHairAndBeardLeaveTheEyesVisible(t *testing.T) {
	r := loadAdult(t)
	eye, _ := ParseHex(r.Spec.Fixed[1])
	fs := r.Spec.FrameSize
	eyes := 0
	for _, p := range r.Poses {
		body, err := ReadPNG(r.PartPath("body", "default", p.Parts["body"]))
		if err != nil {
			t.Fatal(err)
		}
		short, err := ReadPNG(r.PartPath("hair", "short", p.Parts["hair"]))
		if err != nil {
			t.Fatal(err)
		}
		type cover struct{ layer, variant string }
		var covers []cover
		for _, v := range layerVariants(r, "hair") {
			covers = append(covers, cover{"hair", v})
		}
		covers = append(covers, cover{"beard", "beard"})
		for _, cv := range covers {
			img, err := ReadPNG(r.PartPath(cv.layer, cv.variant, p.Parts[cv.layer]))
			if err != nil {
				t.Fatal(err)
			}
			for y := 0; y < fs; y++ {
				for x := 0; x < fs; x++ {
					if c, ok := rgbAt(body, x, y); !ok || c != eye {
						continue
					}
					if _, hidden := rgbAt(short, x, y); hidden {
						continue
					}
					eyes++
					if _, covered := rgbAt(img, x, y); covered {
						t.Errorf("pose %s: %s %s covers an eye at (%d,%d)", p.Name, cv.layer, cv.variant, x, y)
					}
				}
			}
		}
	}
	if eyes == 0 {
		t.Fatal("no eye pixels found in the body layer")
	}
}

// The bald body has a whole head: its silhouette covers the short hair's
// face area, so no style leaves holes (eyes included).
func TestAdultBodyHeadHasNoHoles(t *testing.T) {
	r := loadAdult(t)
	fs := r.Spec.FrameSize
	for _, p := range r.Poses {
		body, err := ReadPNG(r.PartPath("body", "default", p.Parts["body"]))
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < fs/2; y++ {
			first, last := -1, -1
			for x := 0; x < fs; x++ {
				if _, ok := rgbAt(body, x, y); ok {
					if first < 0 {
						first = x
					}
					last = x
				}
			}
			for x := first + 1; first >= 0 && x < last; x++ {
				if _, ok := rgbAt(body, x, y); !ok {
					t.Errorf("pose %s: hole in the head at (%d,%d)", p.Name, x, y)
				}
			}
		}
	}
}

// The dress hides the shirt and trousers completely, in every pose.
func TestAdultDressCoversShirtAndTrousers(t *testing.T) {
	r := loadAdult(t)
	fs := r.Spec.FrameSize
	for _, p := range r.Poses {
		dress, err := ReadPNG(r.PartPath("outfit", "dress", p.Parts["outfit"]))
		if err != nil {
			t.Fatal(err)
		}
		for _, layer := range []string{"shirt", "pants"} {
			img, err := ReadPNG(r.PartPath(layer, "default", p.Parts[layer]))
			if err != nil {
				t.Fatal(err)
			}
			for y := 0; y < fs; y++ {
				for x := 0; x < fs; x++ {
					if _, ok := rgbAt(img, x, y); !ok {
						continue
					}
					if _, ok := rgbAt(dress, x, y); !ok {
						t.Errorf("pose %s: %s shows through the dress at (%d,%d)", p.Name, layer, x, y)
					}
				}
			}
		}
	}
}

// Every entity state the child rig animates, the adult rig does too (the
// server and entdefs use the same state names for both sizes).
func TestAdultHasEveryChildState(t *testing.T) {
	adult := loadAdult(t)
	child, err := Load(filepath.Join("..", "..", "..", "res", "sprites", "player"))
	if err != nil {
		t.Fatal(err)
	}
	for state := range child.Clips.States {
		clip, ok := adult.Clips.States[state]
		if !ok {
			t.Errorf("adult rig has no state %q", state)
			continue
		}
		if _, ok := adult.Clips.Clips[clip]; !ok {
			t.Errorf("adult state %q names missing clip %q", state, clip)
		}
	}
	for _, name := range []string{"idleUp", "idleLeft", "idleRight", "bumpDown", "useDown"} {
		if _, ok := adult.Clips.Clips[name]; !ok {
			t.Errorf("adult rig has no %s clip (the client plays it by name)", name)
		}
	}
}

// Bald, the body, shirt and trousers join up: no empty row between the top
// of the head and the feet down the centre column (e.g. a missing neck).
func TestAdultBaldFigureHasNoGaps(t *testing.T) {
	r := loadAdult(t)
	fs := r.Spec.FrameSize
	cx := r.Spec.AnchorOrDefault()[0]
	for _, p := range r.Poses {
		filled := make([]bool, fs)
		for _, layer := range []string{"body", "shirt", "pants"} {
			img, err := ReadPNG(r.PartPath(layer, "default", p.Parts[layer]))
			if err != nil {
				t.Fatal(err)
			}
			for y := 0; y < fs; y++ {
				for x := cx - 2; x <= cx+2; x++ {
					if _, ok := rgbAt(img, x, y); ok {
						filled[y] = true
					}
				}
			}
		}
		top, bottom := -1, -1
		for y, f := range filled {
			if f {
				if top < 0 {
					top = y
				}
				bottom = y
			}
		}
		for y := top; top >= 0 && y <= bottom; y++ {
			if !filled[y] {
				t.Errorf("pose %s: gap in the figure at row %d", p.Name, y)
			}
		}
	}
}

// The idle breathes: four frames from the standing pose, the second pair with
// only the head and shoulders lowered; `stand` stays still (mannequins).
func TestAdultBreathingIdle(t *testing.T) {
	r := loadAdult(t)
	for _, f := range []string{"Down", "Up", "Left", "Right"} {
		name := "idle" + f
		if f == "Down" {
			name = "idle"
		}
		steps := r.Clips.Clips[name]
		if len(steps) != 4 {
			t.Fatalf("%s: want 4 breathing frames, got %d", name, len(steps))
		}
		for _, s := range steps {
			if s.Ms < 400 {
				t.Errorf("%s: %dms per frame is a pant, not a breath", name, s.Ms)
			}
		}
	}
	stand := r.Clips.Clips[r.Clips.States["stand"]]
	if len(stand) != 1 {
		t.Fatalf("stand state: want one still frame, got %v", stand)
	}
	// Frame 0 is the standing pose; frame 2 differs only in the upper body.
	fs := r.Spec.FrameSize
	for _, d := range []string{"south", "east", "north", "west"} {
		for _, layer := range []string{"body", "shirt", "pants"} {
			standImg, err := ReadPNG(r.PartPath(layer, "default", "idle_"+d+"_0"))
			if err != nil {
				t.Fatal(err)
			}
			in, _ := ReadPNG(r.PartPath(layer, "default", "breathe_"+d+"_0"))
			out, _ := ReadPNG(r.PartPath(layer, "default", "breathe_"+d+"_2"))
			for y := 0; y < fs; y++ {
				for x := 0; x < fs; x++ {
					a, aok := rgbAt(standImg, x, y)
					b, bok := rgbAt(in, x, y)
					if a != b || aok != bok {
						t.Fatalf("%s %s: breathe frame 0 differs from the standing pose at (%d,%d)", layer, d, x, y)
					}
					c, cok := rgbAt(out, x, y)
					if y > 22 && (a != c || aok != cok) {
						t.Fatalf("%s %s: the out-breath moved row %d (only head and shoulders should)", layer, d, y)
					}
				}
			}
		}
	}
}

// The mannequin form is faceless: no eye pixels, no skin colours.
func TestAdultMannequinFormIsFaceless(t *testing.T) {
	r := loadAdult(t)
	eye, _ := ParseHex(r.Spec.Fixed[1])
	skin := map[RGB]bool{}
	for _, k := range r.Spec.Slots["skin"] {
		c, _ := ParseHex(k)
		skin[c] = true
	}
	fs := r.Spec.FrameSize
	for _, p := range r.Poses {
		img, err := ReadPNG(r.PartPath("form", "mannequin", p.Parts["form"]))
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				if c, ok := rgbAt(img, x, y); ok && (c == eye || skin[c]) {
					t.Fatalf("pose %s: mannequin form has a face/skin pixel at (%d,%d)", p.Name, x, y)
				}
			}
		}
	}
}

// Hands in pockets: a second breathing stance (clips pockets*, state
// `pockets`) that differs from the natural idle only around the hands.
func TestAdultHandsInPocketsStance(t *testing.T) {
	r := loadAdult(t)
	if r.Clips.States["pockets"] != "pockets" {
		t.Fatalf("no pockets state: %v", r.Clips.States)
	}
	for _, c := range []string{"pockets", "pocketsUp", "pocketsLeft", "pocketsRight"} {
		if len(r.Clips.Clips[c]) != 4 {
			t.Errorf("%s: want 4 breathing frames, got %d", c, len(r.Clips.Clips[c]))
		}
	}
	idle, _ := ReadPNG(r.PartPath("body", "default", "breathe_south_0"))
	pockets, _ := ReadPNG(r.PartPath("body", "default", "pockets_south_0"))
	fs := r.Spec.FrameSize
	diff := 0
	for y := 0; y < fs; y++ {
		for x := 0; x < fs; x++ {
			a, aok := rgbAt(idle, x, y)
			b, bok := rgbAt(pockets, x, y)
			if a != b || aok != bok {
				diff++
				if y < 20 {
					t.Fatalf("pockets differs from the idle above the hands at (%d,%d)", x, y)
				}
			}
		}
	}
	if diff == 0 {
		t.Fatal("pockets stance is identical to the natural idle")
	}
}
