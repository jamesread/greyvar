package gridgen

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/jamesread/greyvar/datlib/tiled"
	"github.com/jamesread/greyvar/datlib/worldfiles"
)

const repoTileset = "../../../res/img/textures/tilesets/grass_terrains.tsx"

func testConfig(t *testing.T) *GenerationConfig {
	return &GenerationConfig{
		WorldName:    "gentest",
		Title:        "Gen Test",
		OutDir:       t.TempDir(),
		TilesetPath:  repoTileset,
		WangSet:      "terrains",
		Terrains:     []string{"water", "sand", "grass"},
		Thresholds:   []float64{-0.1, 0.05},
		MapRows:      16,
		MapCols:      16,
		WorldRows:    2,
		WorldCols:    2,
		Seed:         1347,
		NoiseScale:   10,
		SpawnTerrain: "grass",
		Entities:     []EntityPlacement{{Definition: "pot", Terrain: "grass", PerMap: 2}},
	}
}

func TestGenerateLoadsAsTiledWorld(t *testing.T) {
	cfg := testConfig(t)
	res, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}

	world, err := worldfiles.LoadWorldFromDir(cfg.WorldName, cfg.OutDir)
	if err != nil {
		t.Fatalf("generated world does not load: %v", err)
	}

	blank, err := tiled.BlankTiles(repoTileset)
	if err != nil {
		t.Fatal(err)
	}

	if len(world.Grids) != 4 {
		t.Fatalf("want 4 grids, got %d", len(world.Grids))
	}
	if world.SpawnGrid != res.SpawnGrid {
		t.Fatalf("spawn grid %q, want %q", world.SpawnGrid, res.SpawnGrid)
	}

	for name, g := range world.Grids {
		if g.RowCount != 16 || g.ColCount != 16 {
			t.Errorf("%s: size %dx%d", name, g.RowCount, g.ColCount)
		}
		if len(g.Layers) != 4 { // three terrain layers + entities
			t.Errorf("%s: %d layers", name, len(g.Layers))
		}
		for r := uint32(0); r < g.RowCount; r++ {
			for c := uint32(0); c < g.ColCount; c++ {
				if tile := g.TopTileAt(r, c); tile == nil || tile.AtlasKey == "" {
					t.Fatalf("%s: cell %d,%d has no tileset tile: %+v", name, r, c, tile)
				}
				for _, layer := range g.Layers {
					if tile := layer.TileAt(r, c); tile != nil && tile.AtlasKey != "" && blank[int(tile.FrameIndex)] {
						t.Fatalf("%s: cell %d,%d uses blank tile %d", name, r, c, tile.FrameIndex)
					}
				}
			}
		}
	}

	if len(world.Grids[res.SpawnGrid].SpawnPoints) != 1 {
		t.Errorf("spawn grid should have one spawn marker")
	}

	if dx, dy, ok := worldfiles.ScrollDeltaBetween(world.Definition, world.Grids, "0.0.tmj", "0.1.tmj"); !ok || dx != 256 || dy != 0 {
		t.Errorf("0.0 -> 0.1 scroll delta = %d,%d,%v", dx, dy, ok)
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	a, b := testConfig(t), testConfig(t)
	if _, err := Generate(a); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(b); err != nil {
		t.Fatal(err)
	}

	wa, _ := worldfiles.LoadWorldFromDir(a.WorldName, a.OutDir)
	wb, _ := worldfiles.LoadWorldFromDir(b.WorldName, b.OutDir)
	for name, ga := range wa.Grids {
		gb := wb.Grids[name]
		for li := range ga.Layers {
			for r, row := range ga.Layers[li].Tiles {
				for c, ta := range row {
					tb := gb.Layers[li].TileAt(r, c)
					if tb == nil || ta.FrameIndex != tb.FrameIndex {
						t.Fatalf("%s layer %d differs at %d,%d", name, li, r, c)
					}
				}
			}
		}
	}
}

// Every emitted pattern must exist in the wangset after fix-up.
func TestFixUnsupportedCoversAllPatterns(t *testing.T) {
	ws, err := tiled.LoadWangSet(repoTileset, "terrains")
	if err != nil {
		t.Fatal(err)
	}
	p, err := LoadPalette(ws, []string{"water", "sand", "grass"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	f := newCornerField(40, 40)
	rng := rand.New(rand.NewSource(7))
	for i := range f.level {
		f.level[i] = rng.Intn(3) // worst case: pure noise
	}
	fixUnsupported(f, p)

	for r := 0; r < 40; r++ {
		for c := 0; c < 40; c++ {
			for level := 1; level < p.Levels(); level++ {
				if mask := maskAtLeast(f.cellCorners(r, c), level); !p.Has(level, mask) {
					t.Fatalf("cell %d,%d level %d mask %s unsupported", r, c, level, maskString(mask))
				}
			}
		}
	}
}

func TestLoadPaletteRejectsUnknownTerrain(t *testing.T) {
	ws, err := tiled.LoadWangSet(repoTileset, "terrains")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPalette(ws, []string{"water", "lava"}, nil); err == nil {
		t.Fatal("expected error for unknown terrain")
	}
	// water has no transparent edge tiles, so it cannot be an upper layer.
	if _, err := LoadPalette(ws, []string{"grass", "water"}, nil); err == nil {
		t.Fatal("expected error layering water over grass")
	}
}

// Variants are weighted by Tiled probability: grass_terrains.tsx gives the
// plain grass tiles 0.05 and the old baked-flower tiles 0 (flowers now come
// from the decor layer, see TestGenerateScattersDecor).
func TestTileHonoursProbability(t *testing.T) {
	ws, err := tiled.LoadWangSet(repoTileset, "terrains")
	if err != nil {
		t.Fatal(err)
	}
	p, err := LoadPalette(ws, []string{"water", "sand", "grass"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	weights := ws.Probabilities()
	ids := p.overlay[2][fullMask]
	total := 0.0
	for _, id := range ids {
		total += weights[id]
	}

	rng := rand.New(rand.NewSource(1))
	counts := map[int]int{}
	const n = 20000
	for i := 0; i < n; i++ {
		id, _ := p.Tile(2, fullMask, rng)
		counts[id]++
	}

	for _, id := range ids {
		want := weights[id] / total
		got := float64(counts[id]) / n
		if got < want-0.02 || got > want+0.02 {
			t.Errorf("tile %d picked %.3f of the time, want ~%.3f", id, got, want)
		}
	}
}

func TestGenerateScattersDecor(t *testing.T) {
	cfg := testConfig(t)
	cfg.DecorTilesetPath = filepath.Join(filepath.Dir(repoTileset), "flora.tsx")
	cfg.DecorDensity = 0.5
	if _, err := Generate(cfg); err != nil {
		t.Fatal(err)
	}
	decor, err := loadDecorTileset(cfg.DecorTilesetPath)
	if err != nil {
		t.Fatal(err)
	}
	grassTiles, err := tsxTileCount(repoTileset)
	if err != nil {
		t.Fatal(err)
	}
	floraFirst := 1 + grassTiles // flora follows every grass_terrains tile (incl. generated water frames)
	isDecorGID := map[uint32]bool{}
	for _, id := range decor.IDs {
		isDecorGID[uint32(floraFirst+id)] = true
	}
	baked := map[uint32]bool{127: true, 128: true, 129: true} // gids of tiles 126-128
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
		for _, l := range doc.Layers {
			for _, g := range l.Data {
				if g == 0 {
					continue
				}
				if l.Name == "decor" {
					if !isDecorGID[g] {
						t.Fatalf("%s: decor layer has non-decor gid %d", name, g)
					}
					placed++
				} else if baked[g] {
					t.Fatalf("%s layer %s: baked flower tile gid %d still generated", name, l.Name, g)
				}
			}
		}
		if placed > 0 && doc.Tilesets[len(doc.Tilesets)-1].FirstGID != floraFirst {
			t.Fatalf("%s: flora tileset should start at gid %d: %+v", name, floraFirst, doc.Tilesets)
		}
	}
	if placed == 0 {
		t.Fatal("expected flowers on the decor layers")
	}
}
