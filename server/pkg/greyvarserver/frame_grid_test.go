package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/server/pkg/worlds"
)

func TestTileVisibleToClient(t *testing.T) {
	if tileVisibleToClient(nil) {
		t.Fatalf("nil tile should not be visible")
	}
	if !tileVisibleToClient(&gridfiles.Tile{}) {
		t.Fatalf("default tile should be visible")
	}
	if tileVisibleToClient(&gridfiles.Tile{Hidden: true}) {
		t.Fatalf("hidden tile should not be visible to client")
	}
}

func TestGenerateGridUpdateSkipsHiddenTiles(t *testing.T) {
	s := &serverInterface{
		loadedWorlds: map[string]*worlds.World{
			"test": {
				Grids: map[string]*gridfiles.Grid{
					"0.0.tmj": {
						Filename: "0.0.tmj",
						RowCount: 2,
						ColCount: 2,
						Layers: []gridfiles.TileLayer{
							gridfiles.NewTileLayer("terrain", gridfiles.LayerKindTerrain),
						},
					},
				},
			},
		},
	}

	grid := s.loadedWorlds["test"].Grids["0.0.tmj"]
	grid.Layers[0].SetTile(0, 0, &gridfiles.Tile{Row: 0, Col: 0, Texture: "grass.png"})
	grid.Layers[0].SetTile(0, 1, &gridfiles.Tile{Row: 0, Col: 1, Texture: "spawn.png", Hidden: true})

	p := &RemotePlayer{
		CurrentWorldId: "test",
		CurrentGridId:  "0.0.tmj",
		NeedsGridUpdate: true,
	}

	update := generateGridUpdate(s, p)
	if update == nil {
		t.Fatalf("expected grid update")
	}
	if len(update.Layers) != 1 {
		t.Fatalf("expected one layer, got %d", len(update.Layers))
	}
	if len(update.Layers[0].Tiles) != 1 {
		t.Fatalf("expected one visible tile, got %d", len(update.Layers[0].Tiles))
	}
	if update.Layers[0].Tiles[0].Tex != "grass.png" {
		t.Fatalf("unexpected visible tile %q", update.Layers[0].Tiles[0].Tex)
	}
}
