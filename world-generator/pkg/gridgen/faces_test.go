package gridgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesread/greyvar/datlib/tiled"
)

// Castle walls are drawn as the top of a taller face; the lower half goes
// in the cell below, on a generated layer directly under the wall layer.
func TestWallFacesGetTheirLowerHalf(t *testing.T) {
	cfg := testConfig(t)
	cfg.Terrains = []string{"grass", "wall"}
	cfg.Thresholds = []float64{0.05}
	cfg.Entities = nil
	if _, err := Generate(cfg); err != nil {
		t.Fatal(err)
	}
	faces, err := tiled.FacesBelow(repoTileset)
	if err != nil || len(faces) == 0 {
		t.Fatalf("tileset has no face_below tiles: %v", err)
	}
	placed := 0
	for _, name := range []string{"0.0.tmj", "0.1.tmj", "1.0.tmj", "1.1.tmj"} {
		raw, err := os.ReadFile(filepath.Join(cfg.OutDir, name))
		if err != nil {
			t.Fatal(err)
		}
		var doc tmjDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		var wall, lower []uint32
		for i, l := range doc.Layers {
			if l.Name == "wall faces (generated)" {
				if i+1 >= len(doc.Layers) || doc.Layers[i+1].Name != "wall" {
					t.Fatalf("%s: the faces layer must sit directly under the wall layer", name)
				}
				lower = l.Data
			}
			if l.Name == "wall" {
				wall = l.Data
			}
		}
		for cell, g := range wall {
			fb, ok := faces[int(g)-1]
			if g == 0 || !ok || cell+16 >= len(wall) {
				continue
			}
			if lower == nil {
				t.Fatalf("%s: wall face at cell %d but no faces layer", name, cell)
			}
			if got := int(lower[cell+16]) - 1; got != fb.Full && got != fb.Short {
				t.Fatalf("%s: under face tile %d at cell %d got %d, want %d or %d", name, g-1, cell, got, fb.Full, fb.Short)
			}
			placed++
		}
	}
	if placed == 0 {
		t.Fatal("no wall faces in the generated world; pick another seed")
	}
}
