package rigs

import (
	"path/filepath"
	"strings"
	"testing"
)

func playerRig(t *testing.T) *Rig {
	t.Helper()
	// The real generated manifest, so this also guards the server/client contract.
	r, err := Load(filepath.Join("..", "..", "res", "img", "textures", "entities", "player", "player.rig.json"))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestDefaultIsStableAndValid(t *testing.T) {
	r := playerRig(t)
	a := r.Default("player7")
	b := r.Default("player7")
	for slot, p := range a.Palettes {
		if b.Palettes[slot] != p {
			t.Fatalf("default not stable for %s", slot)
		}
		if _, ok := r.Palettes[slot][p]; !ok {
			t.Fatalf("default picked unknown palette %s.%s", slot, p)
		}
	}
	if len(a.Palettes) != len(r.SlotOrder) {
		t.Fatalf("want a palette per slot, got %v", a.Palettes)
	}
	distinct := map[string]bool{}
	for _, seed := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		d := r.Default(seed)
		distinct[d.Palettes["shirt"]+d.Palettes["hair"]] = true
	}
	if len(distinct) < 3 {
		t.Fatalf("defaults barely vary across seeds: %v", distinct)
	}
}

func TestApply(t *testing.T) {
	r := playerRig(t)
	base := r.Default("x")
	got, err := r.Apply(base, nil, map[string]string{"hair": "blonde"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Palettes["hair"] != "blonde" || got.Palettes["skin"] != base.Palettes["skin"] {
		t.Fatalf("apply wrong: %v", got.Palettes)
	}
	if base.Palettes["hair"] == "blonde" && base.Palettes["hair"] != got.Palettes["hair"] {
		t.Fatal("base mutated")
	}
	for _, tc := range []struct {
		parts, pals map[string]string
		want        string
	}{
		{nil, map[string]string{"hair": "purple"}, `unknown hair palette "purple"`},
		{nil, map[string]string{"wings": "x"}, `unknown colour slot "wings"`},
		{map[string]string{"hair": "mohawk"}, nil, `unknown hair variant "mohawk"`},
		{map[string]string{"cape": "red"}, nil, `unknown layer "cape"`},
	} {
		if _, err := r.Apply(base, tc.parts, tc.pals); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("want %q, got %v", tc.want, err)
		}
	}
}
