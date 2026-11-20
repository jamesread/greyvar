package rig

import (
	"path/filepath"
	"testing"
)

// On the real player rig, long hair never covers a face pixel that short
// hair leaves visible (eyes, cheeks): in front and side views it frames or
// trails the face; only back views (no head part) are fully covered.
func TestLongHairLeavesTheFaceVisible(t *testing.T) {
	r, err := Load(filepath.Join("..", "..", "..", "res", "sprites", "player"))
	if err != nil {
		t.Fatal(err)
	}
	isSkin := map[RGB]bool{}
	for _, k := range r.Spec.Slots["skin"] {
		c, _ := ParseHex(k)
		isSkin[c] = true
	}
	fs := r.Spec.FrameSize
	checked := 0
	for _, p := range r.Poses {
		headPart, hasHead := p.Parts["head"]
		hairPart, hasHair := p.Parts["hair"]
		if !hasHead || !hasHair {
			continue
		}
		head, err := ReadPNG(r.PartPath("head", "default", headPart))
		if err != nil {
			t.Fatal(err)
		}
		short, err := ReadPNG(r.PartPath("hair", "short", hairPart))
		if err != nil {
			t.Fatal(err)
		}
		long, err := ReadPNG(r.PartPath("hair", "long", hairPart))
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < fs; y++ {
			for x := 0; x < fs; x++ {
				c, skin := rgbAt(head, x, y)
				if !skin || !isSkin[c] {
					continue
				}
				_, shortCovers := rgbAt(short, x, y)
				_, longCovers := rgbAt(long, x, y)
				if longCovers && !shortCovers {
					t.Errorf("pose %s: long hair covers the face at (%d,%d)", p.Name, x, y)
				}
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no poses with a face checked")
	}
}
