package rig

import (
	"strings"
	"testing"
)

func TestColourRef(t *testing.T) {
	r := loadMini(t)
	c, ok, err := r.ColourRef("hair.1")
	if err != nil || !ok || c.Hex() != "#603020" {
		t.Fatalf("hair.1 = %v %v %v", c.Hex(), ok, err)
	}
	if _, ok, err := r.ColourRef("clear"); ok || err != nil {
		t.Fatalf("clear should be transparent")
	}
	for _, bad := range []string{"hair.9", "nope.0", "hair", "fixed.x"} {
		if _, _, err := r.ColourRef(bad); err == nil {
			t.Errorf("%s: want error", bad)
		}
	}
}

func TestDeriveAndShowPart(t *testing.T) {
	r := loadMini(t)
	e1, _ := ParsePixelEdit("7,8=fixed.0")
	e2, _ := ParsePixelEdit("6-9,4=clear")
	if _, err := r.DerivePart("body", "default", "stand", "blink", []PixelEdit{e1, e2}, false); err != nil {
		t.Fatal(err)
	}
	img, err := ReadPNG(r.PartPath("body", "default", "blink"))
	if err != nil {
		t.Fatal(err)
	}
	if c, _ := rgbAt(img, 7, 8); c.Hex() != "#101010" {
		t.Errorf("edit not applied: %s", c.Hex())
	}
	for x := 6; x <= 9; x++ {
		if _, ok := rgbAt(img, x, 4); ok {
			t.Errorf("(%d,4) should be cleared", x)
		}
	}
	if _, err := r.DerivePart("body", "default", "stand", "blink", nil, false); err == nil {
		t.Error("overwrite without force should fail")
	}
	bad, _ := ParsePixelEdit("16,0=hair.0")
	if _, err := r.DerivePart("body", "default", "", "x", []PixelEdit{bad}, false); err == nil {
		t.Error("out-of-bounds edit should fail")
	}
	out, err := r.ShowPart("body", "default", "blink")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"fixed.0", "skin.1"} {
		if !strings.Contains(out, want) {
			t.Errorf("show legend missing %s:\n%s", want, out)
		}
	}
}

func TestDeriveLongHair(t *testing.T) {
	r := loadMini(t)
	paths, err := r.DeriveLongHair(LongHairOptions{Layer: "hair", From: "short", To: "longer", Rows: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 {
		t.Fatalf("want 1 part (stand), got %v", paths)
	}
	img, err := ReadPNG(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	// Mini hair spans x 5..10 down to row 3; outer columns grow to row 5,
	// outlined at row 6 and on the outside.
	for _, x := range []int{5, 10} {
		if _, ok := rgbAt(img, x, 5); !ok {
			t.Errorf("column %d not extended", x)
		}
		if c, _ := rgbAt(img, x, 6); c.Hex() != "#101010" {
			t.Errorf("column %d missing bottom outline, got %s", x, c.Hex())
		}
	}
	if c, _ := rgbAt(img, 4, 4); c.Hex() != "#101010" {
		t.Errorf("missing outer side outline, got %s", c.Hex())
	}
	// Mini poses have no head layer, so they count as back views and every
	// hair column grows (front views only grow the outer two per side).
	if _, ok := rgbAt(img, 7, 5); !ok {
		t.Error("back view: middle column should grow too")
	}
	if _, err := r.DeriveLongHair(LongHairOptions{Layer: "hair", From: "short", To: "longer"}); err == nil {
		t.Error("second run without force should refuse to overwrite")
	}
}

func TestUnusedParts(t *testing.T) {
	r := loadMini(t)
	if u, err := r.UnusedParts(); err != nil || len(u) != 0 {
		t.Fatalf("fresh mini rig: want no unused parts, got %v %v", u, err)
	}
	if _, err := r.DerivePart("body", "default", "stand", "orphan", nil, false); err != nil {
		t.Fatal(err)
	}
	u, err := r.UnusedParts()
	if err != nil || len(u) != 1 || !strings.HasSuffix(u[0], "body/default/orphan.png") {
		t.Fatalf("want orphan listed, got %v %v", u, err)
	}
}
