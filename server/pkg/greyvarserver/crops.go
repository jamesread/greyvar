package greyvarserver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jamesread/greyvar/datlib/common"
	"github.com/jamesread/greyvar/datlib/cropdefs"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

// Crops are server-owned cells: planted from seeds, grown by timestamps (so
// they keep growing while the server is off), harvested when ripe, saved per
// world under the state dir, and drawn by clients as a "crops" grid layer of
// animated tiles from the generated crops tileset.

const cropsLayer = "crops"

// cropSaveEveryMs bounds how often a changed world is written to disk.
const cropSaveEveryMs = 1000

type cropKey struct {
	Grid     string
	Row, Col uint32
}

type cropCell struct {
	Crop    string `json:"crop"`
	Stage   int    `json:"stage"`
	StageAt int64  `json:"stageAt"` // unix ms the current stage began
}

type cropStore struct {
	defs       map[string]*cropdefs.CropDefinition
	seedToCrop map[string]string
	stageTiles map[string]map[int]int
	atlas      *tiled.AtlasTileset
	stateDir   string

	worlds   map[string]map[cropKey]*cropCell
	dirty    map[string]bool
	lastSave int64
}

func newCropStore() *cropStore {
	return &cropStore{
		defs:       map[string]*cropdefs.CropDefinition{},
		seedToCrop: map[string]string{},
		worlds:     map[string]map[cropKey]*cropCell{},
		dirty:      map[string]bool{},
	}
}

// loadCrops reads crop definitions and the crops tileset. Crops are disabled
// (with a warning) if either is missing.
func (s *serverInterface) loadCrops() {
	s.crops = newCropStore()
	s.crops.stateDir = cropStateDir()

	defs, err := cropdefs.LoadAll(filepath.Join(common.DatDir(), "cropdefs"))
	if err != nil {
		log.Warnf("Cannot load cropdefs: %v", err)
		return
	}
	res := common.ResDir()
	if res == "" || len(defs) == 0 {
		return
	}
	tsx := filepath.Join(res, "img", "textures", "tilesets", "crops.tsx")
	cat, _, err := tiled.LoadTilesetFile(tsx)
	if err != nil {
		log.Warnf("Cannot load crops tileset %s: %v", tsx, err)
		return
	}
	if atlases := cat.AtlasTilesets(); len(atlases) == 1 {
		s.crops.atlas = &atlases[0]
	}
	if s.crops.stageTiles, err = cropdefs.StageTiles(tsx); err != nil {
		log.Warnf("Cannot read crop stages from %s: %v", tsx, err)
		return
	}
	for id, def := range defs {
		errs := def.Validate(func(item string) bool { _, ok := s.itemDefinitions[item]; return ok })
		if missing := cropdefs.MissingStageTiles(def, s.crops.stageTiles); len(missing) > 0 {
			errs = append(errs, fmt.Sprintf("stages %v have no tile in crops.tsx", missing))
		}
		if len(errs) > 0 {
			log.Warnf("cropdef %s disabled: %v", id, errs)
			continue
		}
		s.crops.defs[id] = def
		s.crops.seedToCrop[def.Seed] = id
	}
	log.Infof("Loaded %d crop(s); state dir %s", len(s.crops.defs), s.crops.stateDir)
}

func cropStateDir() string {
	if dir := os.Getenv("GREYVAR_STATE_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(common.DatDir(), "..", "state")
}

func (c *cropStore) enabled() bool { return c != nil && c.atlas != nil && len(c.defs) > 0 }

func (c *cropStore) path(worldId string) string {
	return filepath.Join(c.stateDir, "crops", worldId+".json")
}

type cropFileCell struct {
	Grid string `json:"grid"`
	Row  uint32 `json:"row"`
	Col  uint32 `json:"col"`
	cropCell
}

// world returns a world's crops, loading the saved state the first time.
func (c *cropStore) world(worldId string) map[cropKey]*cropCell {
	if cells, ok := c.worlds[worldId]; ok {
		return cells
	}
	cells := map[cropKey]*cropCell{}
	c.worlds[worldId] = cells
	data, err := os.ReadFile(c.path(worldId))
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Cannot read crops for %s: %v", worldId, err)
		}
		return cells
	}
	var saved struct {
		Cells []cropFileCell `json:"cells"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		log.Warnf("Ignoring corrupt crop state %s: %v", c.path(worldId), err)
		return cells
	}
	for _, sc := range saved.Cells {
		if _, ok := c.defs[sc.Crop]; !ok {
			continue // crop type removed since the save
		}
		cell := sc.cropCell
		cells[cropKey{sc.Grid, sc.Row, sc.Col}] = &cell
	}
	log.Infof("Loaded %d crop cell(s) for world %s", len(cells), worldId)
	return cells
}

// save writes one world's crops atomically (temp file + rename).
func (c *cropStore) save(worldId string) error {
	cells := c.worlds[worldId]
	out := struct {
		Cells []cropFileCell `json:"cells"`
	}{Cells: []cropFileCell{}}
	for k, cell := range cells {
		out.Cells = append(out.Cells, cropFileCell{Grid: k.Grid, Row: k.Row, Col: k.Col, cropCell: *cell})
	}
	sort.Slice(out.Cells, func(i, j int) bool {
		a, b := out.Cells[i], out.Cells[j]
		if a.Grid != b.Grid {
			return a.Grid < b.Grid
		}
		if a.Row != b.Row {
			return a.Row < b.Row
		}
		return a.Col < b.Col
	})
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	path := c.path(worldId)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// cropNetTile is the client tile for a crop cell's current stage.
func (c *cropStore) netTile(k cropKey, cell *cropCell) *pb.Tile {
	id := c.stageTiles[cell.Crop][cell.Stage]
	return &pb.Tile{
		Row:              k.Row,
		Col:              k.Col,
		Tex:              fmt.Sprintf("%s#%d", filepath.Base(c.atlas.ImagePath), id),
		AtlasKey:         c.atlas.Key,
		FrameIndex:       int32(id),
		CollisionFromTsx: true,
	}
}

// addCropsToGridUpdate adds the crops tileset and a "crops" layer holding this
// grid's crop cells.
func (s *serverInterface) addCropsToGridUpdate(g *pb.Grid, worldId, gridId string) {
	c := s.crops
	if !c.enabled() {
		return
	}
	g.Tilesets = append(g.Tilesets, &pb.Tileset{
		Key:        c.atlas.Key,
		Image:      c.atlas.ImagePath,
		TileWidth:  uint32(c.atlas.TileWidth),
		TileHeight: uint32(c.atlas.TileHeight),
		Columns:    uint32(c.atlas.Columns),
		Animations: tileAnimationsToPB(c.atlas.Animations, c.atlas.Synced),
	})
	layer := &pb.GridLayer{Name: cropsLayer, Kind: "crops"}
	keys := make([]cropKey, 0)
	for k := range c.world(worldId) {
		if k.Grid == gridId {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Row != keys[j].Row {
			return keys[i].Row < keys[j].Row
		}
		return keys[i].Col < keys[j].Col
	})
	for _, k := range keys {
		layer.Tiles = append(layer.Tiles, c.netTile(k, c.world(worldId)[k]))
	}
	g.Layers = append(g.Layers, layer)
}

// broadcastCrop tells players on that grid about a crop cell (nil = cleared).
func (s *serverInterface) broadcastCrop(worldId string, k cropKey, cell *cropCell) {
	change := &pb.TileChange{Layer: cropsLayer, Row: k.Row, Col: k.Col}
	if cell != nil {
		change.Tile = s.crops.netTile(k, cell)
	}
	for _, p := range s.remotePlayers {
		if p.CurrentWorldId == worldId && p.CurrentGridId == k.Grid {
			p.PendingTileChanges = append(p.PendingTileChanges, change)
		}
	}
	s.crops.dirty[worldId] = true
}

func frameTileChanges(s *serverInterface, p *RemotePlayer) {
	if len(p.PendingTileChanges) == 0 {
		return
	}
	p.currentFrame.TileChanges = append(p.currentFrame.TileChanges, p.PendingTileChanges...)
	p.PendingTileChanges = nil
}

// tickCrops advances growth by timestamp and saves changed worlds.
func (s *serverInterface) tickCrops(nowMs int64) {
	c := s.crops
	if !c.enabled() {
		return
	}
	for worldId, cells := range c.worlds {
		for k, cell := range cells {
			if advanceCrop(c.defs[cell.Crop], cell, nowMs) {
				s.broadcastCrop(worldId, k, cell)
			}
		}
	}
	if nowMs-c.lastSave < cropSaveEveryMs {
		return
	}
	for worldId := range c.dirty {
		if err := c.save(worldId); err != nil {
			log.Warnf("Cannot save crops for %s: %v", worldId, err)
			continue
		}
		delete(c.dirty, worldId)
	}
	c.lastSave = nowMs
}

// advanceCrop moves a cell through every stage whose time has passed.
func advanceCrop(def *cropdefs.CropDefinition, cell *cropCell, nowMs int64) bool {
	if def == nil {
		return false
	}
	changed := false
	for cell.Stage < def.RipeStage() {
		grow := def.Stages[cell.Stage].GrowMs
		if nowMs < cell.StageAt+grow {
			break
		}
		cell.StageAt += grow
		cell.Stage++
		changed = true
	}
	return changed
}

func nowMs() int64 { return time.Now().UnixMilli() }

func (s *serverInterface) sayToPlayer(rp *RemotePlayer, text string) {
	rp.PendingHudMessage = &text
	rp.PendingHudEntityId = rp.Entity.ServerId
}

// cellInReach: the cell centre is within the interact range of the player's
// centre, the same reach as entities.
func cellInReach(player *Entity, row, col uint32) bool {
	px, py := player.X+entitySizePx/2, player.Y+entitySizePx/2
	cx, cy := int32(col)*tileSizePx+tileSizePx/2, int32(row)*tileSizePx+tileSizePx/2
	return absInt32(px-cx) <= interactRangePx && absInt32(py-cy) <= interactRangePx
}

// processCellInteract harvests a ripe crop or plants the held seed in the
// faced cell. Returns whether anything happened.
func (s *serverInterface) processCellInteract(rp *RemotePlayer, row, col uint32) bool {
	c := s.crops
	if !c.enabled() || rp.Entity == nil {
		return false
	}
	grid := s.gridForPlayer(rp)
	if grid == nil || row >= grid.RowCount || col >= grid.ColCount || !cellInReach(rp.Entity, row, col) {
		return false
	}
	worldId := rp.CurrentWorldId
	k := cropKey{rp.CurrentGridId, row, col}
	cells := c.world(worldId)

	if cell, ok := cells[k]; ok {
		def := c.defs[cell.Crop]
		advanceCrop(def, cell, nowMs())
		if cell.Stage < def.RipeStage() {
			s.sayToPlayer(rp, fmt.Sprintf("The %s is still growing.", def.Title))
			return true
		}
		rp.addInventoryItem(def.Harvest.Item, def.Harvest.Count)
		if def.Harvest.Seeds > 0 {
			rp.addInventoryItem(def.Seed, def.Harvest.Seeds)
		}
		rp.PendingInventoryUpdate = true
		delete(cells, k)
		s.broadcastCrop(worldId, k, nil)
		s.sayToPlayer(rp, fmt.Sprintf("Harvested %s x%d", itemTitle(s, def.Harvest.Item), def.Harvest.Count))
		return true
	}

	held := ""
	if rp.Entity.Appearance != nil {
		held = rp.Entity.Appearance.HeldItem
	}
	cropId, isSeed := c.seedToCrop[held]
	if !isSeed {
		return false
	}
	if !s.cellPlantable(rp, row, col) {
		s.sayToPlayer(rp, "Nothing will grow there.")
		return true
	}
	if !rp.removeInventoryItem(held, 1) {
		return false
	}
	rp.PendingInventoryUpdate = true
	cell := &cropCell{Crop: cropId, Stage: 0, StageAt: nowMs()}
	cells[k] = cell
	s.broadcastCrop(worldId, k, cell)
	s.sayToPlayer(rp, "Planted "+itemTitle(s, held))
	return true
}

// cellPlantable: walkable ground (the tile under any decor has no collision)
// with no entity standing on it.
func (s *serverInterface) cellPlantable(rp *RemotePlayer, row, col uint32) bool {
	grid := s.gridForPlayer(rp)
	if grid == nil {
		return false
	}
	if tile := grid.CollisionTileAt(row, col); !isTileTraversable(tile) {
		return false
	}
	for _, ent := range s.entitiesOnGrid(rp.CurrentWorldId, rp.CurrentGridId) {
		if ent.Definition == "player" {
			continue
		}
		er, ec := tileCoordsFromPixels(ent.X+entitySizePx/2, ent.Y+entitySizePx/2)
		if er == row && ec == col {
			return false
		}
	}
	return true
}

// giveStarterSeeds hands new players each crop's starter seeds.
func (s *serverInterface) giveStarterSeeds(rp *RemotePlayer) {
	if !s.crops.enabled() {
		return
	}
	ids := make([]string, 0, len(s.crops.defs))
	for id := range s.crops.defs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if def := s.crops.defs[id]; def.StarterSeeds > 0 {
			rp.addInventoryItem(def.Seed, def.StarterSeeds)
		}
	}
}
