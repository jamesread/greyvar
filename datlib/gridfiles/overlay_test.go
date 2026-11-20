package gridfiles

import "testing"

func TestCollisionTileAtMergesOverlays(t *testing.T) {
	g := &Grid{RowCount: 1, ColCount: 3}
	rim := &Tile{Texture: "rim", Collision: []CollisionRect{{X: 6, W: 4, H: 16}}, CollisionFromTSX: true}
	sand := &Tile{Texture: "sand", CollisionFromTSX: true}
	face := &Tile{Texture: "face", Overlay: true, Collision: []CollisionRect{{X: 10, W: 6, H: 12}}, CollisionFromTSX: true}
	ground := TileLayer{Tiles: map[uint32]map[uint32]*Tile{0: {0: rim, 1: sand}}}
	over := TileLayer{Tiles: map[uint32]map[uint32]*Tile{0: {0: face, 1: face, 2: face}}}
	g.Layers = []TileLayer{ground, over}

	got := g.CollisionTileAt(0, 0)
	if got.Texture != "rim" || len(got.Collision) != 2 {
		t.Fatalf("overlay should add its rect to the rim's, got %+v", got)
	}
	if len(rim.Collision) != 1 {
		t.Fatal("merging must not modify the grid's own tile")
	}
	if got := g.CollisionTileAt(0, 1); got.Texture != "sand" || len(got.Collision) != 1 {
		t.Fatalf("walkable sand under a face should block with the face's rect, got %+v", got)
	}
	if got := g.CollisionTileAt(0, 2); len(got.Collision) != 1 || got.Traversable {
		t.Fatalf("overlay-only cell should keep the overlay's collision, got %+v", got)
	}
	if got := g.TopTileAt(0, 1); got != face {
		t.Fatal("TopTileAt still returns the drawn overlay")
	}
}

func TestCollisionTileAtAddsOverlaysBelow(t *testing.T) {
	// A tall wall face's lower half sits on a layer under the wall layer; a
	// nearer wall in the same cell decides, but the face still collides.
	g := &Grid{RowCount: 1, ColCount: 2}
	grass := &Tile{Texture: "grass", CollisionFromTSX: true}
	lower := &Tile{Texture: "lower face", Overlay: true, Collision: []CollisionRect{{X: 10, W: 6, H: 16}}, CollisionFromTSX: true}
	wall := &Tile{Texture: "wall", Collision: []CollisionRect{{X: 6, W: 4, H: 16}}, CollisionFromTSX: true}
	g.Layers = []TileLayer{
		{Tiles: map[uint32]map[uint32]*Tile{0: {0: grass, 1: grass}}},
		{Tiles: map[uint32]map[uint32]*Tile{0: {0: lower, 1: lower}}},
		{Tiles: map[uint32]map[uint32]*Tile{0: {0: wall}}},
	}
	if got := g.CollisionTileAt(0, 0); got.Texture != "wall" || len(got.Collision) != 2 {
		t.Fatalf("the face under the wall should add its rect, got %+v", got)
	}
	if got := g.CollisionTileAt(0, 1); got.Texture != "grass" || len(got.Collision) != 1 {
		t.Fatalf("the face over open grass should block, got %+v", got)
	}
}
