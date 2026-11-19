package gridgen

import (
	"fmt"
	"strconv"
	"strings"
)

// GenerationConfig describes a generated Tiled world. All tile choices come
// from the wangset in TilesetPath; nothing about textures is hard-coded.
type GenerationConfig struct {
	WorldName string // world directory and .world file stem
	Title     string
	Author    string
	OutDir    string // directory the .world and .tmj files are written to

	TilesetPath string // .tsx with a corner wangset
	WangSet     string
	EntdefsDir  string // used to warn about unknown entity definitions

	// Terrains are wangset color names ordered bottom to top. The first is the
	// base fill; each later one is drawn on its own layer over the ones below,
	// using the set's transparent ("no terrain") transition tiles.
	Terrains []string
	// Thresholds split noise values between terrains: len(Terrains)-1 values,
	// ascending. Noise is roughly in [-1, 1], mostly within [-0.5, 0.5].
	Thresholds []float64

	MapRows   int // tiles per map
	MapCols   int
	WorldRows int // maps in the world
	WorldCols int

	Seed       int64
	NoiseScale float64 // tiles per noise unit; larger = broader features

	SpawnTerrain string // terrain the player spawn marker is placed on
	Entities     []EntityPlacement

	// Decor scatters tiles marked decor=true (flowers) from DecorTilesetPath
	// on a "decor" layer over all-DecorTerrain cells, each cell with
	// probability DecorDensity. Empty path disables it.
	DecorTilesetPath string
	DecorTerrain     string
	DecorDensity     float64
}

// EntityPlacement scatters an entity definition over cells of one terrain.
type EntityPlacement struct {
	Definition string
	Terrain    string
	PerMap     int
}

func (c *GenerationConfig) Validate() error {
	if c.WorldName == "" {
		return fmt.Errorf("world name is required")
	}
	if len(c.Terrains) < 1 {
		return fmt.Errorf("at least one terrain is required")
	}
	if len(c.Thresholds) != len(c.Terrains)-1 {
		return fmt.Errorf("need %d thresholds for %d terrains, got %d", len(c.Terrains)-1, len(c.Terrains), len(c.Thresholds))
	}
	for i := 1; i < len(c.Thresholds); i++ {
		if c.Thresholds[i] <= c.Thresholds[i-1] {
			return fmt.Errorf("thresholds must be ascending: %v", c.Thresholds)
		}
	}
	if c.MapRows < 2 || c.MapCols < 2 || c.WorldRows < 1 || c.WorldCols < 1 {
		return fmt.Errorf("invalid size: maps %dx%d tiles, world %dx%d maps", c.MapRows, c.MapCols, c.WorldRows, c.WorldCols)
	}
	if c.NoiseScale <= 0 {
		return fmt.Errorf("noise scale must be positive")
	}
	if c.DecorDensity < 0 || c.DecorDensity > 1 {
		return fmt.Errorf("decor density must be between 0 and 1, got %v", c.DecorDensity)
	}
	return nil
}

// ParseThresholds parses "a,b,c" into floats.
func ParseThresholds(raw string) ([]float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, fmt.Errorf("threshold %q: %w", p, err)
		}
		out = append(out, v)
	}
	return out, nil
}

// ParseEntityPlacement parses "definition:terrain:perMap".
func ParseEntityPlacement(raw string) (EntityPlacement, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 3 {
		return EntityPlacement{}, fmt.Errorf("entity %q: want definition:terrain:perMap", raw)
	}
	n, err := strconv.Atoi(parts[2])
	if err != nil || n < 0 {
		return EntityPlacement{}, fmt.Errorf("entity %q: perMap must be a non-negative integer", raw)
	}
	return EntityPlacement{Definition: parts[0], Terrain: parts[1], PerMap: n}, nil
}
