package main

import (
	"flag"
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/jamesread/greyvar/world-generator/pkg/puzzlegen"
)

// runPuzzle is `world-generator puzzle ...`: generate (or just preview) one
// verified puzzle world from the difficulty curve.
func runPuzzle(args []string) {
	fs := flag.NewFlagSet("puzzle", flag.ExitOnError)
	var (
		level     = fs.Int("level", 1, "difficulty level (1 = one lever; 9+ = endless, seed-varied)")
		seed      = fs.Int64("seed", 1, "layout seed")
		mechanics = fs.String("mechanics", "", "comma-separated mechanics to allow (default all): lever,keyChest,pushBlock,timedSwitch")
		name      = fs.String("name", "", "world name (default puzzle_<level>_<seed>)")
		worldsDir = fs.String("worlds-dir", "../server/dat/worlds", "directory that holds world folders")
		outDir    = fs.String("out", "", "output directory (overrides -worlds-dir/-name)")
		tileset   = fs.String("tileset", "../res/img/textures/tilesets/grass_terrains.tsx", "terrain tileset (wangset \"terrains\" with grass and wall)")
		decorTs   = fs.String("decor-tileset", "../res/img/textures/tilesets/flora.tsx", "decor tileset; empty disables")
		entdefs   = fs.String("entdefs", "../server/dat/entdefs", "entity definitions dir")
		preview   = fs.Bool("preview", false, "print the puzzle and its solution without writing files")
	)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "usage: world-generator puzzle [flags]")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	allowed, err := puzzlegen.ParseMechanics(splitList(*mechanics))
	if err != nil {
		fatal(err)
	}
	if *mechanics == "" {
		allowed = nil // the curve's own mix
	}
	spec := puzzlegen.Curve(*level, *seed, allowed)

	if *preview {
		res, err := puzzlegen.Build(spec, puzzlegen.Options{})
		if err != nil {
			fatal(err)
		}
		fmt.Print(puzzlegen.Describe(res))
		return
	}

	if *name == "" {
		*name = fmt.Sprintf("puzzle_%d_%d", *level, *seed)
	}
	dir := *outDir
	if dir == "" {
		dir = filepath.Join(*worldsDir, *name+"_tiled")
	}
	res, err := puzzlegen.Generate(spec, puzzlegen.Options{
		WorldName:        *name,
		OutDir:           dir,
		TilesetPath:      *tileset,
		DecorTilesetPath: *decorTs,
		EntdefsDir:       *entdefs,
	})
	if err != nil {
		fatal(err)
	}
	fmt.Print(puzzlegen.Describe(res))
	log.Infof("wrote %s (spawn grid %s)", res.WorldPath, res.SpawnGrid)
	log.Infof("load it with GREYVAR_WORLD=%s", filepath.Base(dir))
}
