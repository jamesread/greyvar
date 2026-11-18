package tiled

import (
	"os"
	"path/filepath"
	"testing"
)

const wangTSX = `<?xml version="1.0" encoding="UTF-8"?>
<tileset version="1.10" name="t" tilewidth="16" tileheight="16" tilecount="4" columns="2">
 <image source="t.png" width="32" height="32"/>
 <tile id="3" probability="0.25"/>
 <wangsets>
  <wangset name="terrains" type="corner" tile="-1">
   <wangcolor name="sand" color="#ff7700" tile="-1" probability="1"/>
   <wangcolor name="grass" color="#24ff20" tile="-1" probability="0.5"/>
   <wangtile tileid="0" wangid="0,1,0,1,0,1,0,1"/>
   <wangtile tileid="1" wangid="0,2,0,2,0,2,0,2"/>
   <wangtile tileid="2" wangid="0,2,0,2,0,1,0,1"/>
   <wangtile tileid="3" wangid="0,2,0,2,0,1,0,1"/>
  </wangset>
  <wangset name="paths" type="edge" tile="-1">
   <wangcolor name="path" color="#ff0000" tile="-1" probability="1"/>
  </wangset>
 </wangsets>
</tileset>`

func writeWangTSX(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "t.tsx")
	if err := os.WriteFile(path, []byte(wangTSX), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadWangSets(t *testing.T) {
	sets, err := LoadWangSets(writeWangTSX(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 2 {
		t.Fatalf("want 2 wangsets, got %d", len(sets))
	}

	ws := sets[0]
	if ws.Name != "terrains" || ws.Type != "corner" || len(ws.Colors) != 2 || len(ws.Tiles) != 4 {
		t.Fatalf("unexpected wangset: %+v", ws)
	}
	if ws.Colors[1].Probability != 0.5 {
		t.Fatalf("probability not parsed: %+v", ws.Colors[1])
	}

	if idx, ok := ws.ColorIndex("Grass"); !ok || idx != 2 {
		t.Fatalf("ColorIndex(Grass) = %d, %v", idx, ok)
	}
	if _, ok := ws.ColorIndex("lava"); ok {
		t.Fatal("ColorIndex(lava) should not exist")
	}
}

func TestWangTileCorners(t *testing.T) {
	ws, err := LoadWangSet(writeWangTSX(t), "terrains")
	if err != nil {
		t.Fatal(err)
	}

	// wangid 0,2,0,2,0,1,0,1 -> NE=2, SE=2, SW=1, NW=1
	byCorners := ws.TilesByCorners()
	variants := byCorners[Corners{1, 2, 2, 1}]
	if len(variants) != 2 || variants[0] != 2 || variants[1] != 3 {
		t.Fatalf("variants for grass-east pattern = %v", variants)
	}
	if got := byCorners[Corners{2, 2, 2, 2}]; len(got) != 1 || got[0] != 1 {
		t.Fatalf("full grass = %v", got)
	}

	probs := ws.Probabilities()
	if probs[3] != 0.25 || probs[2] != 1 {
		t.Fatalf("probabilities = %v, want tile 3 = 0.25 and default 1 for tile 2", probs)
	}
}

func TestLoadWangSetMissing(t *testing.T) {
	if _, err := LoadWangSet(writeWangTSX(t), "nope"); err == nil {
		t.Fatal("expected error for missing wangset")
	}
}

func TestParseWangIDRejectsLegacyFormat(t *testing.T) {
	if _, err := parseWangID("0x10101010"); err == nil {
		t.Fatal("expected error for legacy hex wangid")
	}
}

func TestBlankTilesOnRepoTileset(t *testing.T) {
	blank, err := BlankTiles("../../res/img/textures/tilesets/grass_terrains.tsx")
	if err != nil {
		t.Skipf("repo tileset unavailable: %v", err)
	}
	if blank[153] {
		t.Fatal("tile 153 (water) has pixels and must not be blank")
	}
	if !blank[0] {
		t.Fatal("tile 0 is empty in grass_terrains.png and should be blank")
	}
}
