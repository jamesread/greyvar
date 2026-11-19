package gridfiles

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/tiled"
)

// left half of a tile (x 0-7)
func leftHalf() *tiled.PixelMask {
	var m tiled.PixelMask
	for y := range m {
		m[y] = 0x00ff
	}
	return &m
}

func oneCell(layers ...*Tile) *Grid {
	g := &Grid{RowCount: 1, ColCount: 1}
	for _, t := range layers {
		l := NewTileLayer("l", LayerKindTerrain)
		l.Tiles = map[uint32]map[uint32]*Tile{0: {0: t}}
		g.Layers = append(g.Layers, l)
	}
	return g
}

func TestSeeThroughEdgeBlocksOnlyOverWater(t *testing.T) {
	water := &Tile{Water: true, Collision: []CollisionRect{{W: 16, H: 16}}}
	edge := &Tile{Texture: "sand-edge", SeeThrough: leftHalf()}
	applyWaterCollision(oneCell(water, edge))
	if len(edge.Collision) != 1 || edge.Collision[0] != (CollisionRect{X: 0, Y: 0, W: 8, H: 16}) {
		t.Fatalf("over water the see-through half should block: %+v", edge.Collision)
	}

	grass := &Tile{Texture: "grass"}
	edge2 := &Tile{Texture: "sand-edge", SeeThrough: leftHalf()}
	applyWaterCollision(oneCell(grass, edge2))
	if len(edge2.Collision) != 0 {
		t.Fatalf("over grass it must stay walkable: %+v", edge2.Collision)
	}
	if len(water.Collision) != 1 {
		t.Fatal("water tiles keep their own collision only")
	}
}

func TestWaterColouredPixelsAlwaysBlock(t *testing.T) {
	lake := &Tile{Texture: "lake-edge", WaterPixels: leftHalf()}
	applyWaterCollision(oneCell(lake))
	if len(lake.Collision) != 1 || lake.Collision[0].W != 8 {
		t.Fatalf("lake edge's water should block: %+v", lake.Collision)
	}
}

func TestSeeThroughStackAndDecor(t *testing.T) {
	// water <- grass edge (left half see-through) <- flower (decor) on top:
	// the decor is ignored and the edge's see-through half shows water.
	water := &Tile{Water: true}
	edge := &Tile{Texture: "grass-edge", SeeThrough: leftHalf()}
	flower := &Tile{Texture: "flower", Decor: true}
	applyWaterCollision(oneCell(water, edge, flower))
	if len(edge.Collision) != 1 || len(flower.Collision) != 0 {
		t.Fatalf("edge %+v flower %+v", edge.Collision, flower.Collision)
	}
	// Two see-through layers: only where both are clear does water show.
	var top tiled.PixelMask
	for y := range top {
		top[y] = 0xffff
	}
	mid := &Tile{Texture: "mid", SeeThrough: leftHalf()}
	upper := &Tile{Texture: "upper", SeeThrough: &top}
	applyWaterCollision(oneCell(&Tile{Water: true}, mid, upper))
	if len(upper.Collision) != 1 || upper.Collision[0].W != 8 {
		t.Fatalf("water shows only through both: %+v", upper.Collision)
	}
}

func TestMaskRectsSnapAndMerge(t *testing.T) {
	var m tiled.PixelMask
	m[0] = 0x0001 // a single stray pixel: ignored
	for y := 8; y < 16; y++ {
		m[y] = 0xffff
	}
	got := maskRects(m)
	if len(got) != 1 || got[0] != (CollisionRect{X: 0, Y: 8, W: 16, H: 8}) {
		t.Fatalf("want the bottom half as one rect, got %+v", got)
	}
}
