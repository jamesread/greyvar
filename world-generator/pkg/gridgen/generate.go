package gridgen

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/jamesread/greyvar/datlib/tiled"
	"github.com/jamesread/greyvar/datlib/worldfiles"
)

// Result summarises a generated world.
type Result struct {
	WorldPath     string
	Maps          []string
	SpawnGrid     string
	CornersRaised int
	CellsByLevel  []int // cells whose topmost full terrain is each level
}

type cellPos struct {
	row, col int
}

// Generate builds a Tiled world (.world + one .tmj per map) from noise, using
// the tiles a .tsx wangset declares for each configured terrain.
func Generate(cfg *GenerationConfig) (*Result, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ws, err := tiled.LoadWangSet(cfg.TilesetPath, cfg.WangSet)
	if err != nil {
		return nil, err
	}

	blank, err := tiled.BlankTiles(cfg.TilesetPath)
	if err != nil {
		log.Warnf("cannot check tileset image for blank tiles: %v", err)
	}

	palette, err := LoadPalette(ws, cfg.Terrains, blank)
	if err != nil {
		return nil, err
	}
	faces, err := tiled.FacesBelow(cfg.TilesetPath)
	if err != nil {
		return nil, err
	}
	if len(palette.Stale) > 0 {
		log.Warnf("%s: wangset %q tags tiles %v but they are blank in the image; ignoring them (stale terrain data in the .tsx)",
			filepath.Base(cfg.TilesetPath), cfg.WangSet, palette.Stale)
	}
	for level := 1; level < palette.Levels(); level++ {
		if missing := palette.MissingMasks(level); len(missing) > 0 {
			log.Infof("terrain %q has no tile for corners %v; those patterns will be filled in", cfg.Terrains[level], missing)
		}
	}

	spawnLevel, ok := palette.LevelOf(cfg.SpawnTerrain)
	if !ok {
		return nil, fmt.Errorf("spawn terrain %q is not one of %v", cfg.SpawnTerrain, cfg.Terrains)
	}

	entityLevels := make([]int, len(cfg.Entities))
	for i, ep := range cfg.Entities {
		level, ok := palette.LevelOf(ep.Terrain)
		if !ok {
			return nil, fmt.Errorf("entity %q terrain %q is not one of %v", ep.Definition, ep.Terrain, cfg.Terrains)
		}
		entityLevels[i] = level
		warnUnknownEntdef(cfg.EntdefsDir, ep.Definition)
	}

	tilesetSource, err := relativeSlashPath(cfg.OutDir, cfg.TilesetPath)
	if err != nil {
		return nil, err
	}

	var decor *decorTileset
	var decorSource string
	decorLevel, decorFirst := 0, 0
	if cfg.DecorTilesetPath != "" && cfg.DecorDensity > 0 {
		if decor, err = loadDecorTileset(cfg.DecorTilesetPath); err != nil {
			return nil, err
		}
		terrain := cfg.DecorTerrain
		if terrain == "" {
			terrain = cfg.SpawnTerrain
		}
		var ok bool
		if decorLevel, ok = palette.LevelOf(terrain); !ok {
			return nil, fmt.Errorf("decor terrain %q is not one of %v", terrain, cfg.Terrains)
		}
		count, err := tsxTileCount(cfg.TilesetPath)
		if err != nil {
			return nil, err
		}
		decorFirst = 1 + count
		if decorSource, err = relativeSlashPath(cfg.OutDir, cfg.DecorTilesetPath); err != nil {
			return nil, err
		}
	}

	var field *cornerField
	if cfg.Field != nil {
		field = cfg.Field.Clone()
		for i, v := range field.level {
			if v < 0 || v >= palette.Levels() {
				return nil, fmt.Errorf("field corner %d has level %d; terrains are 0..%d", i, v, palette.Levels()-1)
			}
		}
	} else {
		field = newCornerField(cfg.WorldRows*cfg.MapRows, cfg.WorldCols*cfg.MapCols)
		fillFromNoise(field, cfg)
	}

	res := &Result{CellsByLevel: make([]int, palette.Levels())}
	res.CornersRaised = fixUnsupported(field, palette)

	rng := rand.New(rand.NewSource(cfg.Seed))
	spawnMapRow, spawnMapCol := cfg.WorldRows/2, cfg.WorldCols/2
	res.SpawnGrid = mapFilename(spawnMapRow, spawnMapCol)

	var placements []worldfiles.MapPlacement

	for wr := 0; wr < cfg.WorldRows; wr++ {
		for wc := 0; wc < cfg.WorldCols; wc++ {
			name := mapFilename(wr, wc)
			m := &mapBuilder{
				cfg:     cfg,
				field:   field,
				palette: palette,
				rng:     rng,
				row0:    wr * cfg.MapRows,
				col0:    wc * cfg.MapCols,
				faces:   faces,
			}

			doc := m.build(tilesetSource, res.CellsByLevel)
			used := map[cellPos]bool{}

			for _, o := range cfg.Objects[name] {
				p := cellPos{o.Row, o.Col}
				used[p] = true
				m.addObjectWithProps(o.Type, o.Name, p, o.Props)
			}

			if wr == spawnMapRow && wc == spawnMapCol && !cfg.NoAutoSpawn {
				if pos, ok := m.spawnCell(spawnLevel); ok {
					used[pos] = true
					m.addObject("spawn", "spawn", pos)
				} else {
					log.Warnf("%s has no all-%s cell for a spawn marker; the server will use the map centre", name, cfg.SpawnTerrain)
				}
			}

			for i, ep := range cfg.Entities {
				placed := m.scatter(ep.Definition, entityLevels[i], ep.PerMap, used)
				if placed < ep.PerMap {
					log.Warnf("%s: placed %d/%d %s (not enough all-%s cells)", name, placed, ep.PerMap, ep.Definition, ep.Terrain)
				}
			}

			if decor != nil {
				if data := m.decorLayer(decorLevel, cfg.DecorDensity, decor, decorFirst, used); data != nil {
					doc.Layers = append(doc.Layers, tmjLayer{
						ID: len(doc.Layers) + 1, Type: "tilelayer", Name: "decor",
						Width: cfg.MapCols, Height: cfg.MapRows, Data: data, Opacity: 1, Visible: true,
					})
					doc.Tilesets = append(doc.Tilesets, tmjTileset{FirstGID: decorFirst, Source: decorSource})
				}
			}

			doc.Layers = append(doc.Layers, m.objectLayer(len(doc.Layers)+1))
			doc.NextLayerID = len(doc.Layers) + 1
			doc.NextObjectID = len(m.objects) + 1

			if err := writeJSON(filepath.Join(cfg.OutDir, name), doc); err != nil {
				return nil, fmt.Errorf("write %s: %w", name, err)
			}

			res.Maps = append(res.Maps, name)
			placements = append(placements, worldfiles.MapPlacement{
				FileName: name,
				Width:    cfg.MapCols * tileSize,
				Height:   cfg.MapRows * tileSize,
				X:        wc * cfg.MapCols * tileSize,
				Y:        wr * cfg.MapRows * tileSize,
			})
		}
	}

	world := worldDoc{
		Type:                 "world",
		Maps:                 placements,
		OnlyShowAdjacentMaps: false,
		Properties: []tiledProperty{
			{Name: "author", Type: "string", Value: cfg.Author},
			{Name: "title", Type: "string", Value: cfg.Title},
			{Name: "spawnGrid", Type: "string", Value: res.SpawnGrid},
			{Name: "seed", Type: "string", Value: fmt.Sprintf("%d", cfg.Seed)},
		},
	}
	keys := make([]string, 0, len(cfg.WorldProperties))
	for k := range cfg.WorldProperties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		world.Properties = append(world.Properties, tiledProperty{Name: k, Type: "string", Value: cfg.WorldProperties[k]})
	}

	res.WorldPath = filepath.Join(cfg.OutDir, cfg.WorldName+".world")
	if err := writeJSON(res.WorldPath, world); err != nil {
		return nil, fmt.Errorf("write world: %w", err)
	}

	return res, nil
}

func mapFilename(row, col int) string {
	return fmt.Sprintf("%d.%d.tmj", row, col)
}

func relativeSlashPath(fromDir, target string) (string, error) {
	absFrom, err := filepath.Abs(fromDir)
	if err != nil {
		return "", err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absFrom, absTarget)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func warnUnknownEntdef(dir, definition string) {
	if dir == "" {
		return
	}
	if _, err := os.Stat(filepath.Join(dir, definition+".yml")); err != nil {
		log.Warnf("entity definition %q not found in %s", definition, dir)
	}
}

// mapBuilder renders one map's slice of the world corner field.
type mapBuilder struct {
	cfg        *GenerationConfig
	field      *cornerField
	palette    *Palette
	rng        *rand.Rand
	row0, col0 int
	objects    []tmjObject
	faces      map[int]tiled.FaceBelow // tile -> lower half of its tall south face
}

func (m *mapBuilder) corners(row, col int) [4]int {
	return m.field.cellCorners(m.row0+row, m.col0+col)
}

func (m *mapBuilder) build(tilesetSource string, cellsByLevel []int) *tmjDoc {
	rows, cols := m.cfg.MapRows, m.cfg.MapCols
	levels := m.palette.Levels()

	doc := &tmjDoc{
		CompressionLevel: -1,
		Width:            cols,
		Height:           rows,
		Orientation:      "orthogonal",
		RenderOrder:      "right-down",
		TiledVersion:     "1.12.2",
		Type:             "map",
		Version:          "1.10",
		TileWidth:        tileSize,
		TileHeight:       tileSize,
		Tilesets:         []tmjTileset{{FirstGID: 1, Source: tilesetSource}},
	}

	for level := 0; level < levels; level++ {
		data := make([]uint32, rows*cols)
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				corners := m.corners(r, c)
				mask := uint8(fullMask)
				if level > 0 {
					mask = maskAtLeast(corners, level)
				}
				if mask == 0 {
					continue
				}
				// Skip tiles fully hidden under the next layer's full tile.
				if level+1 < levels && maskAtLeast(corners, level+1) == fullMask {
					continue
				}
				id, ok := m.palette.Tile(level, mask, m.rng)
				if !ok {
					// fixUnsupported guarantees coverage; keep the map loadable regardless.
					log.Errorf("no tile for %s corners %s", m.palette.Terrains[level], maskString(mask))
					continue
				}
				data[r*cols+c] = uint32(id + 1)
			}
		}
		name := m.palette.Terrains[level]
		if level == 0 {
			name = "base: " + name
		}
		doc.Layers = append(doc.Layers, tmjLayer{
			ID:      level + 1,
			Type:    "tilelayer",
			Name:    name,
			Width:   cols,
			Height:  rows,
			Data:    data,
			Opacity: 1,
			Visible: true,
		})
	}

	m.addFaceLayers(doc)

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if level, ok := uniformLevel(m.corners(r, c)); ok {
				cellsByLevel[level]++
			}
		}
	}

	return doc
}

func uniformLevel(corners [4]int) (int, bool) {
	for _, v := range corners[1:] {
		if v != corners[0] {
			return 0, false
		}
	}
	return corners[0], true
}

func (m *mapBuilder) cellsAtLevel(level int) []cellPos {
	var out []cellPos
	for r := 0; r < m.cfg.MapRows; r++ {
		for c := 0; c < m.cfg.MapCols; c++ {
			if l, ok := uniformLevel(m.corners(r, c)); ok && l == level {
				out = append(out, cellPos{r, c})
			}
		}
	}
	return out
}

// spawnCell picks the all-level cell nearest the map centre.
func (m *mapBuilder) spawnCell(level int) (cellPos, bool) {
	cr, cc := m.cfg.MapRows/2, m.cfg.MapCols/2
	best, bestDist, found := cellPos{}, 0, false
	for _, p := range m.cellsAtLevel(level) {
		d := (p.row-cr)*(p.row-cr) + (p.col-cc)*(p.col-cc)
		if !found || d < bestDist {
			best, bestDist, found = p, d, true
		}
	}
	return best, found
}

func (m *mapBuilder) scatter(definition string, level int, count int, used map[cellPos]bool) int {
	candidates := m.cellsAtLevel(level)
	m.rng.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })

	placed := 0
	for _, p := range candidates {
		if placed >= count {
			break
		}
		if used[p] {
			continue
		}
		used[p] = true
		m.addObject(definition, "", p)
		placed++
	}
	return placed
}

func (m *mapBuilder) addObject(typ, name string, p cellPos) {
	m.addObjectWithProps(typ, name, p, nil)
}

func (m *mapBuilder) addObjectWithProps(typ, name string, p cellPos, props map[string]string) {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var tp []tiledProperty
	for _, k := range keys {
		tp = append(tp, tiledProperty{Name: k, Type: "string", Value: props[k]})
	}
	m.objects = append(m.objects, tmjObject{
		Properties: tp,
		ID:         len(m.objects) + 1,
		Name:       name,
		Type:       typ,
		X:          float64(p.col * tileSize),
		Y:          float64(p.row * tileSize),
		Width:      tileSize,
		Height:     tileSize,
		Visible:    true,
	})
}

func (m *mapBuilder) objectLayer(id int) tmjLayer {
	return tmjLayer{
		ID:        id,
		Type:      "objectgroup",
		Name:      "entities",
		DrawOrder: "topdown",
		Objects:   m.objects,
		Opacity:   1,
		Visible:   true,
	}
}

// addFaceLayers gives tall south faces their lower half. A terrain tile with
// a face_below property continues its face into the cell below; that lower
// tile goes on a generated layer directly under the terrain's own layer, so
// a wall standing in that cell (rooms back to back) draws over it. Cells in
// the next map down are left alone (each map is written on its own).
func (m *mapBuilder) addFaceLayers(doc *tmjDoc) {
	if len(m.faces) == 0 {
		return
	}
	rows, cols := m.cfg.MapRows, m.cfg.MapCols
	var out []tmjLayer
	for i, l := range doc.Layers {
		level := i // build adds one layer per level, in order
		var data []uint32
		for r := 0; r+1 < rows; r++ {
			for c := 0; c < cols; c++ {
				g := l.Data[r*cols+c]
				if g == 0 {
					continue
				}
				fb, ok := m.faces[int(g-1)]
				if !ok {
					continue
				}
				id := fb.Full
				if tiled.FaceBelowShort(levelCorners(m.corners(r, c), level), levelCorners(m.corners(r+1, c), level), 1) {
					id = fb.Short
				}
				if data == nil {
					data = make([]uint32, rows*cols)
				}
				data[(r+1)*cols+c] = uint32(id + 1)
			}
		}
		if data != nil {
			out = append(out, tmjLayer{
				Type: "tilelayer", Name: l.Name + " faces (generated)",
				Width: cols, Height: rows, Data: data, Opacity: 1, Visible: true,
			})
		}
		out = append(out, l)
	}
	for i := range out {
		out[i].ID = i + 1
	}
	doc.Layers = out
}

// levelCorners marks each corner 1 when it is at least this level.
func levelCorners(c [4]int, level int) tiled.Corners {
	var out tiled.Corners
	for i, v := range c {
		if v >= level {
			out[i] = 1
		}
	}
	return out
}
