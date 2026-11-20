package rig

import (
	"path/filepath"
	"testing"
)

// On the real player rig, a dress covers every pixel the shirt and trousers
// draw in every pose (with their offsets), so nothing worn underneath ever
// shows; the "none" outfit draws nothing; and every pose names the outfit
// part for the shirt and trousers it actually uses.
func TestDressCoversShirtAndTrousers(t *testing.T) {
	r, err := Load(filepath.Join("..", "..", "..", "res", "sprites", "player"))
	if err != nil {
		t.Fatal(err)
	}
	fs := r.Spec.FrameSize
	placed := func(layer, variant string, p Pose) map[[2]int]bool {
		img, err := ReadPNG(r.PartPath(layer, variant, p.Parts[layer]))
		if err != nil {
			t.Fatal(err)
		}
		off := p.Offsets[layer]
		out := map[[2]int]bool{}
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				if _, ok := rgbAt(img, x, y); ok {
					out[[2]int{x + off[0], y + off[1]}] = true
				}
			}
		}
		return out
	}
	checked := 0
	for _, p := range r.Poses {
		part, ok := p.Parts["outfit"]
		if !ok {
			t.Fatalf("pose %s has no outfit part", p.Name)
		}
		if want := OutfitPartName(p, "shirt", "pants"); part != want {
			t.Fatalf("pose %s: outfit part %q, want %q (run `part outfit -poses`)", p.Name, part, want)
		}
		if len(placed("outfit", "none", p)) != 0 {
			t.Fatalf("pose %s: the none outfit draws pixels", p.Name)
		}
		dress := placed("outfit", "dress", p)
		for _, layer := range []string{"shirt", "pants"} {
			for px := range placed(layer, "default", p) {
				if px[0] < 0 || px[1] < 0 || px[0] >= fs || px[1] >= fs {
					continue
				}
				if !dress[px] {
					t.Fatalf("pose %s: %s pixel %v shows under the dress", p.Name, layer, px)
				}
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no poses checked")
	}
}
