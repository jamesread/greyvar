package worldfiles

import (
	"path/filepath"
	"testing"
)

func fixtureDatDir(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("testdata/dat")
	if err != nil {
		t.Fatalf("resolve fixture dat dir: %v", err)
	}
	return abs
}

func TestLoadWorldYAML(t *testing.T) {
	t.Setenv("GREYVAR_DAT_DIR", fixtureDatDir(t))

	world, err := LoadWorld("yamlWorld_dev")
	if err != nil {
		t.Fatalf("LoadWorld yaml: %v", err)
	}

	if world.Definition.Format != "yaml" {
		t.Fatalf("expected yaml format, got %q", world.Definition.Format)
	}

	if len(world.Grids) != 2 {
		t.Fatalf("expected 2 yaml world grids, got %d", len(world.Grids))
	}

	if world.SpawnGrid != "0.0.grid" {
		t.Fatalf("expected spawn grid 0.0.grid, got %q", world.SpawnGrid)
	}
}

func TestLoadWorldTiled(t *testing.T) {
	t.Setenv("GREYVAR_DAT_DIR", fixtureDatDir(t))

	world, err := LoadWorld("tiledWorld_dev")
	if err != nil {
		t.Fatalf("LoadWorld tiled: %v", err)
	}

	if world.Definition.Format != "tiled-world" {
		t.Fatalf("expected tiled-world format, got %q", world.Definition.Format)
	}

	if len(world.Definition.Maps) != 2 {
		t.Fatalf("expected 2 map placements, got %d", len(world.Definition.Maps))
	}

	if len(world.Grids) != 2 {
		t.Fatalf("expected 2 loaded tmj grids, got %d", len(world.Grids))
	}

	if _, ok := world.Grids["0.1.tmj"]; !ok {
		t.Fatalf("expected 0.1.tmj grid")
	}

	if world.SpawnGrid != "0.0.tmj" {
		t.Fatalf("expected spawn grid 0.0.tmj, got %q", world.SpawnGrid)
	}

	if world.Definition.Properties["foo"] != "bar" {
		t.Fatalf("expected world property foo=bar")
	}
}

func TestListWorldsIncludesTiledWorld(t *testing.T) {
	t.Setenv("GREYVAR_DAT_DIR", fixtureDatDir(t))

	worlds, err := ListWorlds()
	if err != nil {
		t.Fatalf("ListWorlds: %v", err)
	}

	found := false
	for _, summary := range worlds {
		if summary.ID == "tiledWorld_dev" {
			found = true
			if summary.Format != "tiled-world" {
				t.Fatalf("expected tiled-world format in summary")
			}
			if summary.GridCount != 2 {
				t.Fatalf("expected 2 grids in summary, got %d", summary.GridCount)
			}
		}
	}

	if !found {
		t.Fatalf("expected tiledWorld_dev in world list")
	}
}
