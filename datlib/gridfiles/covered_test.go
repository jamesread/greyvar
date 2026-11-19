package gridfiles

import "testing"

func layerWith(name string, kind LayerKind, tiles map[[2]uint32]*Tile) TileLayer {
	l := NewTileLayer(name, kind)
	for rc, t := range tiles {
		l.SetTile(rc[0], rc[1], t)
	}
	return l
}

func TestCoveredDecorAndEntitiesDoNotSpawn(t *testing.T) {
	grass := &Tile{Texture: "grass"}
	hill := &Tile{Texture: "hill"}
	flower := func() *Tile { return &Tile{Texture: "flower", Decor: true} }
	g := &Grid{RowCount: 1, ColCount: 3}
	g.Layers = []TileLayer{
		layerWith("ground", LayerKindTerrain, map[[2]uint32]*Tile{{0, 0}: grass, {0, 1}: grass, {0, 2}: grass}),
		layerWith("decor", LayerKindTerrain, map[[2]uint32]*Tile{{0, 0}: flower(), {0, 1}: flower()}),
		layerWith("low objects", LayerKindObject, map[[2]uint32]*Tile{{0, 1}: {Texture: "rock"}}),
		layerWith("terrain", LayerKindTerrain, map[[2]uint32]*Tile{{0, 1}: hill}),
		layerWith("top decor", LayerKindTerrain, map[[2]uint32]*Tile{{0, 1}: flower()}),
		layerWith("objects", LayerKindObject, nil),
	}
	g.Entities = []GridFileEntityInstance{
		{Row: 0, Col: 1, Definition: "mushroom-under-hill"},
		{Row: 0, Col: 1, Definition: "sign-on-top"},
		{Row: 0, Col: 0, Definition: "mushroom-in-the-open"},
	}
	tiles, ents := dropCovered(g, []int{2, 5, 2})

	if g.Layers[1].TileAt(0, 0) == nil {
		t.Error("a flower in the open stays")
	}
	if g.Layers[1].TileAt(0, 1) != nil {
		t.Error("a flower under the hill should not spawn")
	}
	if g.Layers[2].TileAt(0, 1) != nil {
		t.Error("an object tile under the hill should not spawn")
	}
	if g.Layers[4].TileAt(0, 1) == nil {
		t.Error("a flower drawn on top of the hill stays")
	}
	if g.Layers[0].TileAt(0, 1) != grass || g.Layers[3].TileAt(0, 1) != hill {
		t.Error("ordinary terrain stacking is untouched")
	}
	if tiles != 2 || ents != 1 {
		t.Errorf("dropped %d tiles, %d entities; want 2, 1", tiles, ents)
	}
	if len(g.Entities) != 2 || g.Entities[0].Definition != "sign-on-top" || g.Entities[1].Definition != "mushroom-in-the-open" {
		t.Errorf("entities kept: %+v", g.Entities)
	}
}

func TestDecorDoesNotCoverDecor(t *testing.T) {
	g := &Grid{RowCount: 1, ColCount: 1}
	low := &Tile{Texture: "flower", Decor: true}
	g.Layers = []TileLayer{
		layerWith("a", LayerKindTerrain, map[[2]uint32]*Tile{{0, 0}: low}),
		layerWith("b", LayerKindTerrain, map[[2]uint32]*Tile{{0, 0}: {Texture: "flower2", Decor: true}}),
	}
	dropCovered(g, nil)
	if g.Layers[0].TileAt(0, 0) != low {
		t.Error("a flower above another flower does not hide it")
	}
}
