package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jamesread/greyvar/world-generator/pkg/gridgen"
)

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

func main() {
	if len(os.Args) > 1 && os.Args[1] == "puzzle" {
		runPuzzle(os.Args[2:])
		return
	}

	var (
		name       = flag.String("name", "generated", "world name; output goes to <worlds-dir>/<name>_tiled/<name>.world")
		worldsDir  = flag.String("worlds-dir", "../server/dat/worlds", "directory that holds world folders")
		outDir     = flag.String("out", "", "output directory (overrides -worlds-dir/-name)")
		title      = flag.String("title", "", "world title (default: name)")
		author     = flag.String("author", "world-generator", "world author property")
		tileset    = flag.String("tileset", "../res/img/textures/tilesets/grass_terrains.tsx", ".tsx tileset with a corner wangset")
		wangSet    = flag.String("wangset", "terrains", "wangset name in the tileset")
		entdefs    = flag.String("entdefs", "../server/dat/entdefs", "entity definitions dir, used to sanity-check -entity")
		terrains   = flag.String("terrains", "water,sand,grass", "wangset colors, bottom layer first")
		thresholds = flag.String("thresholds", "-0.1,0.05", "noise cut-offs between terrains (one fewer than -terrains)")
		mapRows    = flag.Int("map-rows", 16, "tiles per map, vertically")
		mapCols    = flag.Int("map-cols", 16, "tiles per map, horizontally")
		worldRows  = flag.Int("world-rows", 2, "maps in the world, vertically")
		worldCols  = flag.Int("world-cols", 2, "maps in the world, horizontally")
		seed       = flag.Int64("seed", 1347, "noise and placement seed")
		scale      = flag.Float64("scale", 10, "noise scale in tiles; larger gives broader terrain")
		spawn      = flag.String("spawn-terrain", "", "terrain for the player spawn marker (default: top terrain)")
		decorTs    = flag.String("decor-tileset", "../res/img/textures/tilesets/flora.tsx", "tileset whose decor=true tiles (flowers) are scattered on a decor layer; empty disables")
		decorTerr  = flag.String("decor-terrain", "", "terrain decor is scattered on (default: spawn terrain)")
		decorDens  = flag.Float64("decor-density", 0.15, "chance each free all-decor-terrain cell gets a decor tile")
		entities   stringList
	)
	flag.Var(&entities, "entity", "scatter entities as definition:terrain:perMap (repeatable)")
	flag.Parse()

	cfg := &gridgen.GenerationConfig{
		WorldName:    *name,
		Title:        *title,
		Author:       *author,
		OutDir:       *outDir,
		TilesetPath:  *tileset,
		WangSet:      *wangSet,
		EntdefsDir:   *entdefs,
		Terrains:     splitList(*terrains),
		MapRows:      *mapRows,
		MapCols:      *mapCols,
		WorldRows:    *worldRows,
		WorldCols:    *worldCols,
		Seed:         *seed,
		NoiseScale:   *scale,
		SpawnTerrain: *spawn,

		DecorTilesetPath: *decorTs,
		DecorTerrain:     *decorTerr,
		DecorDensity:     *decorDens,
	}
	if cfg.OutDir == "" {
		cfg.OutDir = filepath.Join(*worldsDir, *name+"_tiled")
	}
	if cfg.Title == "" {
		cfg.Title = *name
	}
	if cfg.SpawnTerrain == "" && len(cfg.Terrains) > 0 {
		cfg.SpawnTerrain = cfg.Terrains[len(cfg.Terrains)-1]
	}

	var err error
	if cfg.Thresholds, err = gridgen.ParseThresholds(*thresholds); err != nil {
		fatal(err)
	}
	for _, raw := range entities {
		ep, err := gridgen.ParseEntityPlacement(raw)
		if err != nil {
			fatal(err)
		}
		cfg.Entities = append(cfg.Entities, ep)
	}

	res, err := gridgen.Generate(cfg)
	if err != nil {
		fatal(err)
	}

	log.Infof("wrote %s (%d maps, spawn grid %s)", res.WorldPath, len(res.Maps), res.SpawnGrid)
	for level, n := range res.CellsByLevel {
		log.Infof("  %-8s %d solid cells", cfg.Terrains[level], n)
	}
	if res.CornersRaised > 0 {
		log.Infof("  raised %d corners to avoid patterns the tileset has no tile for", res.CornersRaised)
	}
	log.Infof("load it with GREYVAR_WORLD=%s", filepath.Base(cfg.OutDir))
}

func splitList(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "world-generator:", err)
	os.Exit(1)
}
