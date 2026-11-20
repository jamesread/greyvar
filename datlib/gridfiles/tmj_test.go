package gridfiles

import (
	"path/filepath"
	"testing"
)

// Fixtures live in worldfiles/testdata so the tests don't depend on live
// world content under server/dat, which designers edit freely.
const fixtureWorldDir = "../worldfiles/testdata/dat/worlds/tiledWorld_dev"

func TestReadGridTMJ(t *testing.T) {
	path := filepath.Join(fixtureWorldDir, "0.0.tmj")
	grid, err := ReadGridTMJ(path)
	if err != nil {
		t.Fatalf("ReadGridTMJ: %v", err)
	}

	if grid.ColCount != 16 || grid.RowCount != 16 {
		t.Fatalf("expected 16x16 grid, got %dx%d", grid.ColCount, grid.RowCount)
	}

	if grid.Tiles[0][0] == nil || grid.Tiles[0][0].Texture == "" {
		t.Fatalf("expected populated tile at 0,0")
	}

	if len(grid.Entities) == 0 {
		t.Fatalf("expected at least one entity from object layer")
	}

	if len(grid.Tilesets) == 0 {
		t.Fatalf("expected atlas tilesets from grass.tsx")
	}

	foundGrass := false
	for _, tileset := range grid.Tilesets {
		if tileset.Key == "grass" {
			foundGrass = true
			if tileset.ImagePath != "img/textures/tilesets/grass.png" {
				t.Fatalf("expected grass tileset image path, got %q", tileset.ImagePath)
			}
			if tileset.Columns != 20 {
				t.Fatalf("expected grass tileset columns 20, got %d", tileset.Columns)
			}
		}
	}
	if !foundGrass {
		t.Fatalf("expected grass tileset in grid.Tilesets")
	}

	atlasTiles := 0
	highAtlasTiles := 0
	for _, pos := range grid.CellIterator() {
		tile := grid.Tiles[pos.Row][pos.Col]
		if tile.AtlasKey == "grass" {
			atlasTiles++
			if tile.FrameIndex >= 300 {
				highAtlasTiles++
			}
		}
	}
	if atlasTiles == 0 {
		t.Fatalf("expected atlas-backed tiles in grid")
	}
	if highAtlasTiles == 0 {
		t.Fatalf("expected high-index atlas tiles (tilecount fix)")
	}

	foundTV := false
	for _, ent := range grid.Entities {
		if ent.Definition == "tv" {
			foundTV = true
			// A tile object at y=96 is drawn by Tiled from y=80 (its
			// bottom-left is the anchor): row 5.
			if ent.Row != 5 || ent.Col != 6 {
				t.Fatalf("expected tv at row 5 col 6, got row %d col %d", ent.Row, ent.Col)
			}
		}
	}

	if !foundTV {
		t.Fatalf("expected tv entity from tile object gid")
	}
}

func TestReadGridTMJSpawnPoints(t *testing.T) {
	path := filepath.Join(fixtureWorldDir, "0.1.tmj")
	grid, err := ReadGridTMJ(path)
	if err != nil {
		t.Fatalf("ReadGridTMJ: %v", err)
	}

	if len(grid.SpawnPoints) != 1 {
		t.Fatalf("expected 1 spawn point, got %d", len(grid.SpawnPoints))
	}

	var hiddenSpawn *Tile
	for _, layer := range grid.Layers {
		for _, cols := range layer.Tiles {
			for _, tile := range cols {
				if tile != nil && tile.AtlasKey == "gameplay" && tile.FrameIndex == 0 {
					hiddenSpawn = tile
				}
			}
		}
	}
	if hiddenSpawn == nil {
		for _, pos := range grid.CellIterator() {
			tile := grid.TopTileAt(pos.Row, pos.Col)
			if tile != nil && tile.AtlasKey == "gameplay" && tile.FrameIndex == 0 {
				hiddenSpawn = tile
			}
		}
	}
	if hiddenSpawn == nil {
		t.Fatalf("expected gameplay spawn tile on the markers layer")
	}
	if !hiddenSpawn.Hidden {
		t.Fatalf("expected gameplay spawn tile to be hidden when placed on a layer")
	}

	sp := grid.SpawnPoints[0]
	// The spawn tile object at y=128 shows in Tiled at y 112..128: row 7.
	if sp.X != 112 || sp.Y != 112 || sp.Col != 7 || sp.Row != 7 {
		t.Fatalf("spawn point = (%d,%d) row/col (%d,%d), want (112,112) (7,7)", sp.X, sp.Y, sp.Row, sp.Col)
	}

	foundChest := false
	for _, ent := range grid.Entities {
		if ent.Definition == "chest" {
			foundChest = true
			break
		}
	}
	if !foundChest {
		t.Fatalf("expected chest entity from chest.tx template instance (visible omitted)")
	}
}

func TestObjectVisibleDefaultsTrue(t *testing.T) {
	if !objectVisible(tmjMapObject{}) {
		t.Fatal("omitted visible should default to true")
	}
	f := false
	if objectVisible(tmjMapObject{Visible: &f}) {
		t.Fatal("explicit visible:false should be hidden")
	}
	tr := true
	if !objectVisible(tmjMapObject{Visible: &tr}) {
		t.Fatal("explicit visible:true should be shown")
	}
}

func TestReadGridDispatch(t *testing.T) {
	tmjPath := filepath.Join(fixtureWorldDir, "0.0.tmj")
	if _, err := ReadGrid(tmjPath); err != nil {
		t.Fatalf("ReadGrid tmj dispatch: %v", err)
	}

	yamlPath := "../worldfiles/testdata/dat/worlds/yamlWorld_dev/grids/0.0.grid"
	if _, err := ReadGrid(yamlPath); err != nil {
		t.Fatalf("ReadGrid yaml dispatch: %v", err)
	}
}

// Objects land in the cell Tiled shows them in: tile objects are anchored
// bottom-left, others top-left, and off-grid positions go to the nearest
// cell (never a negative one).
func TestObjectCellMatchesTiled(t *testing.T) {
	cases := []struct {
		x, y, w, h float64
		tile       bool
		col, row   uint32
	}{
		{32, 64, 16, 16, true, 2, 3},     // tile object: drawn y 48..64
		{33.25, 66, 16, 16, true, 2, 3},  // dragged a little off the grid
		{11, 86.25, 16, 16, true, 1, 4},  // nearer col 1 than col 0
		{137.5, 111.25, 15, 15, true, 9, 6},
		{15.5, -0.25, 16, 32, false, 1, 0}, // rectangle just above the map's top edge
		{48, 0, 16, 32, false, 3, 0},
		{-3, -3, 16, 16, false, 0, 0},
	}
	for _, c := range cases {
		col, row := objectCell(tmjMapObject{X: c.x, Y: c.y, Width: c.w, Height: c.h}, c.tile, 16, 16)
		if col != c.col || row != c.row {
			t.Errorf("object at (%v,%v) tile=%v: cell (%d,%d), want (%d,%d)", c.x, c.y, c.tile, col, row, c.col, c.row)
		}
	}
}
