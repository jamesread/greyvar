package gridfiles

import "testing"

func TestCollisionTileAtSkipsDecor(t *testing.T) {
	g := &Grid{RowCount: 2, ColCount: 2}
	water := &Tile{Texture: "water", Collision: []CollisionRect{{W: 16, H: 16}}}
	flower := &Tile{Texture: "flora#0", Decor: true}
	ground := TileLayer{Tiles: map[uint32]map[uint32]*Tile{0: {0: water, 1: nil}}}
	decor := TileLayer{Tiles: map[uint32]map[uint32]*Tile{0: {0: flower, 1: flower}}}
	g.Layers = []TileLayer{ground, decor}

	if got := g.TopTileAt(0, 0); got != flower {
		t.Fatal("TopTileAt should still return the drawn (decor) tile")
	}
	if got := g.CollisionTileAt(0, 0); got != water {
		t.Fatalf("CollisionTileAt should see the water under the flower, got %+v", got)
	}
	got := g.CollisionTileAt(0, 1)
	if got == nil || got.Decor || !got.Traversable {
		t.Fatalf("decor-only cell should act empty and traversable, got %+v", got)
	}
}

// isleOfStarting 1.1 after `spritegen flora migrate`: every flower is an
// animated decor tile on top of walkable full grass.
func TestMigratedMapFlowersAreDecor(t *testing.T) {
	g, err := ReadGridTMJ("../../server/dat/worlds/isleOfStarting_dev_tiled/1.1.tmj")
	if err != nil {
		t.Fatal(err)
	}
	flowers := 0
	for _, layer := range g.Layers {
		for _, cols := range layer.Tiles {
			for _, tile := range cols {
				if tile != nil && tile.Decor {
					flowers++
					if layer.Name != "decor" || tile.AtlasKey != "flora" {
						t.Fatalf("decor tile on layer %q atlas %q", layer.Name, tile.AtlasKey)
					}
				}
			}
		}
	}
	// Where a flower is the visible top tile, collision sees the grass below.
	for _, pos := range g.CellIterator() {
		if top := g.TopTileAt(pos.Row, pos.Col); top != nil && top.Decor {
			if under := g.CollisionTileAt(pos.Row, pos.Col); under.Decor || len(under.Collision) > 0 {
				t.Fatalf("cell %v: collision tile should be plain grass, got %+v", pos, under)
			}
		}
	}
	// 82 are painted; 40 sit under sand, path, wall or cliff tiles drawn on
	// a higher layer and do not spawn (covered.go).
	if flowers != 42 {
		t.Fatalf("want 42 visible flowers on 1.1, got %d", flowers)
	}
}

// Water moved from the hard-coded water.png special case to water.tsx: an
// animated tile with a full-tile collision shape, so it still blocks.
func TestRemappedWaterBlocksAndAnimates(t *testing.T) {
	g, err := ReadGridTMJ("../../server/dat/worlds/isleOfStarting_dev_tiled/0.0.tmj")
	if err != nil {
		t.Fatal(err)
	}
	water := 0
	for _, pos := range g.CellIterator() {
		tile := g.CollisionTileAt(pos.Row, pos.Col)
		if tile == nil || tile.AtlasKey != "water" {
			continue
		}
		water++
		if len(tile.Collision) != 1 || tile.Collision[0].W != 16 || tile.Collision[0].H != 16 {
			t.Fatalf("water at %v should have a full-tile collision shape, got %+v", pos, tile.Collision)
		}
	}
	if water == 0 {
		t.Fatal("expected water tiles from water.tsx on isleOfStarting 0.0")
	}
	var animated bool
	for _, ts := range g.Tilesets {
		if ts.Key == "water" && len(ts.Animations[0]) == 3 {
			animated = true
		}
	}
	if !animated {
		t.Fatal("water tileset should animate tile 0 over 3 frames")
	}
}

// See-through sand shores: generated maps put them over water (tide plays),
// hand-made isleOfStarting uses the same tiles over grass (must stay still).
func TestShoreTilesKnowWhatIsBelow(t *testing.T) {
	count := func(path string) (overWater, overOther int) {
		g, err := ReadGridTMJ(path)
		if err != nil {
			t.Fatal(err)
		}
		for li, layer := range g.Layers {
			for row, cols := range layer.Tiles {
				for col, tile := range cols {
					if tile == nil || !tile.NeedsWaterBelow {
						continue
					}
					if g.WaterBelow(li, row, col) {
						overWater++
					} else {
						overOther++
					}
				}
			}
		}
		return
	}
	if w, o := count("../../server/dat/worlds/generated_tiled/1.1.tmj"); w == 0 || o != 0 {
		t.Errorf("generated 1.1: want every shore over water, got %d over water, %d not", w, o)
	}
	if w, o := count("../../server/dat/worlds/isleOfStarting_dev_tiled/1.1.tmj"); o == 0 || w != 0 {
		t.Errorf("isleOfStarting 1.1: want sand-over-grass shores to stay still, got %d over water, %d not", w, o)
	}
	if w, o := count("../../server/dat/worlds/isleOfStarting_dev_tiled/1.0.tmj"); w == 0 || o == 0 {
		t.Errorf("isleOfStarting 1.0 has both kinds: got %d over water, %d not", w, o)
	}
}
