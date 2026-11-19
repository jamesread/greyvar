package faces

import (
	"path/filepath"
	"testing"

	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
)

func build(t *testing.T) *Result {
	t.Helper()
	f, err := Load(filepath.Join("..", "..", "..", "res", "sprites", "faces"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := f.Build()
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func find(t *testing.T, res *Result, tileset string, id int) *FaceTile {
	t.Helper()
	for _, ft := range res.Tiles {
		if filepath.Base(ft.Tileset) == tileset && ft.ID == id {
			return ft
		}
	}
	t.Fatalf("%s#%d is not a face tile", tileset, id)
	return nil
}

func TestFaceTilesFromWangCorners(t *testing.T) {
	res := build(t)
	// South edge, its sand variant, outer corners and notches grow ...
	for _, id := range []int{90, 170, 89, 91, 10, 11} {
		find(t, res, "grass_terrains.tsx", id)
	}
	// ... north/side edges and the plateau itself do not.
	for _, ft := range res.Tiles {
		if filepath.Base(ft.Tileset) == "grass_terrains.tsx" {
			switch ft.ID {
			case 70, 69, 71, 49, 50, 51, 30, 31:
				t.Errorf("tile %d has no south face but was grown", ft.ID)
			}
		}
	}
}

func TestExtensionCollision(t *testing.T) {
	res := build(t)
	if got := find(t, res, "grass_terrains.tsx", 90).ExtCollision; len(got) != 1 || got[0] != [4]int{0, 0, 16, 12} {
		t.Errorf("straight face should block the whole width, 12px down: %v", got)
	}
	// Notch: only its right part has a face (x 9 outline .. 15).
	if got := find(t, res, "grass_terrains.tsx", 10).ExtCollision; len(got) != 1 || got[0][0] != 9 || got[0][2] != 7 {
		t.Errorf("notch face should block x 9-15 only: %v", got)
	}
	// Steps: the stairs stay walkable and are wide enough for the 8px hitbox.
	for _, id := range []int{0, 1} {
		rects := find(t, res, "steps.tsx", id).ExtCollision
		blocked := make([]bool, 16)
		for _, r := range rects {
			for x := r[0]; x < r[0]+r[2]; x++ {
				blocked[x] = true
			}
		}
		run, best := 0, 0
		for _, b := range blocked {
			if b {
				run = 0
			} else if run++; run > best {
				best = run
			}
		}
		if best < 8 {
			t.Errorf("steps#%d: walkable stair gap is %dpx, need >= 8", id, best)
		}
	}
}

func TestBuildIsUpToDate(t *testing.T) {
	if stale := rig.Stale(build(t).Outputs); len(stale) > 0 {
		t.Fatalf("run `make sprites`: %v", stale)
	}
}
