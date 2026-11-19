package greyvarserver

import (
	"strings"
	"testing"

	"github.com/jamesread/greyvar/datlib/cropdefs"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/itemdefs"
	"github.com/jamesread/greyvar/datlib/rigs"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func newCropTestServer(t *testing.T) (*serverInterface, *RemotePlayer, *RemotePlayer) {
	t.Helper()
	blockingTileTextures = map[string]bool{}
	grid := &gridfiles.Grid{RowCount: 3, ColCount: 3}
	grid.Build()
	for r := uint32(0); r < 3; r++ {
		for c := uint32(0); c < 3; c++ {
			grid.Tiles[r][c] = &gridfiles.Tile{Row: r, Col: c, Texture: "grass.png", Traversable: true, CollisionFromTSX: true}
		}
	}
	// A wall at (0,2): nothing grows there.
	grid.Tiles[0][2] = &gridfiles.Tile{Row: 0, Col: 2, Texture: "wall.png", CollisionFromTSX: true,
		Collision: []gridfiles.CollisionRect{{W: 16, H: 16}}}

	s, rp := testServerWithGrid(grid)
	s.itemDefinitions = map[string]*itemdefs.ItemDefinition{"hay": {ID: "hay", Title: "Hay"}, "hay_seeds": {ID: "hay_seeds", Title: "Hay Seeds"}}
	s.crops = newCropStore()
	s.crops.stateDir = t.TempDir()
	s.crops.atlas = &tiled.AtlasTileset{Key: "crops", ImagePath: "img/textures/tilesets/crops.png", TileWidth: 16, TileHeight: 16, Columns: 8}
	s.crops.stageTiles = map[string]map[int]int{"hay": {0: 0, 1: 4, 2: 8}}
	hay := &cropdefs.CropDefinition{ID: "hay", Title: "hay", Seed: "hay_seeds", StarterSeeds: 2,
		Stages: []cropdefs.Stage{{GrowMs: 1000}, {GrowMs: 2000}, {}}, Harvest: cropdefs.Harvest{Item: "hay", Count: 2, Seeds: 1}}
	s.crops.defs["hay"] = hay
	s.crops.seedToCrop["hay_seeds"] = "hay"

	player := &Entity{ServerId: 1, Definition: "player", X: 16, Y: 16, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	rp.Entity = player
	rp.Username = "farmer"
	rp.KnownEntities = map[int64]*Entity{}
	rp.currentFrame = &pb.ServerUpdate{}
	other := &RemotePlayer{Username: "watcher", CurrentWorldId: "testWorld", CurrentGridId: "testGrid", KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"farmer": rp, "watcher": other}

	s.giveStarterSeeds(rp)
	rp.Entity.Appearance = &rigs.Appearance{HeldItem: "hay_seeds"}
	return s, rp, other
}

func TestPlantGrowHarvest(t *testing.T) {
	s, rp, other := newCropTestServer(t)
	if rp.inventoryCount("hay_seeds") != 2 {
		t.Fatalf("starter seeds = %d, want 2", rp.inventoryCount("hay_seeds"))
	}

	// Plant at (1,2), next to the player standing on (1,1).
	if !s.processCellInteract(rp, 1, 2) {
		t.Fatal("planting should be handled")
	}
	if rp.inventoryCount("hay_seeds") != 1 {
		t.Fatal("planting should use one seed")
	}
	k := cropKey{"testGrid", 1, 2}
	cell := s.crops.world("testWorld")[k]
	if cell == nil || cell.Stage != 0 {
		t.Fatalf("expected a stage-0 crop, got %+v", cell)
	}
	if ch := other.PendingTileChanges; len(ch) != 1 || ch[0].Layer != "crops" || ch[0].Tile.GetFrameIndex() != 0 {
		t.Fatalf("other players should see the sprout: %+v", ch)
	}

	// Out of reach, blocked ground and an occupied cell are refused.
	if s.processCellInteract(rp, 2, 3) {
		t.Fatal("out-of-grid cell must be ignored")
	}
	s.processCellInteract(rp, 0, 2)
	if _, planted := s.crops.world("testWorld")[cropKey{"testGrid", 0, 2}]; planted {
		t.Fatal("must not plant on a wall")
	}
	rp.Entity.X, rp.Entity.Y = 0, 0 // (2,2) is now beyond reach
	if s.processCellInteract(rp, 2, 2) {
		t.Fatal("cells beyond reach must be ignored")
	}
	rp.Entity.X, rp.Entity.Y = 16, 16

	// Still growing: interacting just reports it.
	if !s.processCellInteract(rp, 1, 2) || !strings.Contains(*rp.PendingHudMessage, "still growing") {
		t.Fatal("unripe crop should say it is still growing")
	}

	// Growth is by timestamp: jump past both stages at once (server was off).
	cell.StageAt -= 3500
	other.PendingTileChanges = nil
	s.tickCrops(nowMs())
	if cell.Stage != 2 {
		t.Fatalf("crop should be ripe after catching up, stage %d", cell.Stage)
	}
	if ch := other.PendingTileChanges; len(ch) != 1 || ch[0].Tile.GetFrameIndex() != 8 {
		t.Fatalf("ripe tile should be broadcast: %+v", ch)
	}

	// Harvest: hay + a seed back, cell cleared.
	other.PendingTileChanges = nil
	if !s.processCellInteract(rp, 1, 2) {
		t.Fatal("harvest should be handled")
	}
	if rp.inventoryCount("hay") != 2 || rp.inventoryCount("hay_seeds") != 2 {
		t.Fatalf("after harvest: hay=%d seeds=%d, want 2 and 2", rp.inventoryCount("hay"), rp.inventoryCount("hay_seeds"))
	}
	if _, still := s.crops.world("testWorld")[k]; still {
		t.Fatal("harvested cell should be empty")
	}
	if ch := other.PendingTileChanges; len(ch) != 1 || ch[0].Tile != nil {
		t.Fatalf("clear should be broadcast with no tile: %+v", ch)
	}
}

func TestCropsPersistAcrossRestart(t *testing.T) {
	s, rp, _ := newCropTestServer(t)
	s.processCellInteract(rp, 1, 2)
	s.crops.lastSave = 0
	s.tickCrops(nowMs()) // saves the dirty world
	if len(s.crops.dirty) != 0 {
		t.Fatal("world should be saved and clean")
	}

	// A new store over the same state dir sees the crop.
	restarted := newCropStore()
	restarted.stateDir = s.crops.stateDir
	restarted.defs = s.crops.defs
	cell := restarted.world("testWorld")[cropKey{"testGrid", 1, 2}]
	if cell == nil || cell.Crop != "hay" || cell.Stage != 0 {
		t.Fatalf("crop not restored: %+v", cell)
	}
}

func TestGridUpdateIncludesCropsLayer(t *testing.T) {
	s, rp, _ := newCropTestServer(t)
	s.processCellInteract(rp, 1, 2)
	g := &pb.Grid{}
	s.addCropsToGridUpdate(g, "testWorld", "testGrid")
	if len(g.Tilesets) != 1 || g.Tilesets[0].Key != "crops" {
		t.Fatalf("crops tileset missing: %+v", g.Tilesets)
	}
	if len(g.Layers) != 1 || g.Layers[0].Name != "crops" || len(g.Layers[0].Tiles) != 1 {
		t.Fatalf("crops layer missing: %+v", g.Layers)
	}
}

func TestInteractFallsBackToCell(t *testing.T) {
	s, rp, _ := newCropTestServer(t)
	// No entity id (or one with nothing to do): the faced cell is used.
	processInteractRequest(s, rp, &pb.InteractRequest{EntityId: 999, Cell: &pb.CellTarget{Row: 1, Col: 0}})
	if s.crops.world("testWorld")[cropKey{"testGrid", 1, 0}] == nil {
		t.Fatal("interact with no usable entity should plant in the faced cell")
	}
}
