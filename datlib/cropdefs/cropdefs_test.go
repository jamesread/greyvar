package cropdefs

import (
	"path/filepath"
	"testing"
)

func TestRepoHayCropIsComplete(t *testing.T) {
	defs, err := LoadAll(filepath.Join("..", "..", "server", "dat", "cropdefs"))
	if err != nil {
		t.Fatal(err)
	}
	hay := defs["hay"]
	if hay == nil {
		t.Fatal("hay cropdef missing")
	}
	items := map[string]bool{"hay": true, "hay_seeds": true}
	if errs := hay.Validate(func(id string) bool { return items[id] }); len(errs) > 0 {
		t.Fatalf("hay invalid: %v", errs)
	}
	tiles, err := StageTiles(filepath.Join("..", "..", "res", "img", "textures", "tilesets", "crops.tsx"))
	if err != nil {
		t.Fatal(err)
	}
	if m := MissingStageTiles(hay, tiles); len(m) > 0 {
		t.Fatalf("hay stages without tiles: %v", m)
	}
	if hay.RipeStage() != 2 {
		t.Fatalf("ripe stage = %d", hay.RipeStage())
	}
}

func TestValidateCatchesMistakes(t *testing.T) {
	bad := &CropDefinition{ID: "x", Seed: "nope", Stages: []Stage{{GrowMs: 0}, {}}, Harvest: Harvest{Item: "x"}}
	errs := bad.Validate(func(string) bool { return false })
	if len(errs) < 4 {
		t.Fatalf("want growMs, unknown items and count errors, got %v", errs)
	}
}
