package gridfiles

import "github.com/jamesread/greyvar/datlib/tiled"

type LayerKind string

const (
	LayerKindTerrain LayerKind = "terrain"
	LayerKindObject  LayerKind = "object"
)

type Tile struct {
	Row uint32
	Col uint32
	Rot int32
	FlipH bool `yaml:"flipH"`
	FlipV bool `yaml:"flipV"`
	Traversable bool
	Texture string
	AtlasKey string `yaml:"-"`
	FrameIndex int32 `yaml:"-"`
	Collision []CollisionRect
	CollisionFromTSX bool `yaml:"-"`
	TeleportDst string `yaml:"teleportDst,omitempty"`
	TeleportDir string `yaml:"teleportDir,omitempty"`
	TeleportX int `yaml:"teleportX,omitempty"`
	TeleportY int `yaml:"teleportY,omitempty"`
	Hidden bool `yaml:"-"`
	// Decor tiles (tile property decor=true, e.g. flowers) draw on top but
	// are skipped for collision and tile triggers; see CollisionTileAt.
	Decor bool `yaml:"-"`
	// Water: an all-water tile. NeedsWaterBelow: a see-through shore tile
	// whose animation should only play when water is under it.
	Water           bool `yaml:"-"`
	NeedsWaterBelow bool `yaml:"-"`
	// Overlay tiles (property overlay=true, e.g. generated cliff faces and the
	// lower halves of tall wall faces) add their collision to the cell's
	// deciding tile: CollisionTileAt returns that tile with the overlays'
	// rects appended, whether the overlay is drawn above it or below it.
	Overlay bool `yaml:"-"`
	// Above tiles (property above=true, e.g. a tree) draw over characters.
	// For collision they act like overlays: their own rects (the trunk) are
	// added, and they do not hide the ground beneath.
	Above bool `yaml:"-"`
	// SeeThrough / WaterPixels: the tile art's transparent and water-coloured
	// pixels, used to give water collision where it shows (water_collision.go).
	SeeThrough  *tiled.PixelMask `yaml:"-"`
	WaterPixels *tiled.PixelMask `yaml:"-"`
}

// WaterBelow reports whether the tile at (layer, row, col) mostly shows
// water through its see-through pixels (at least half of them), looking down
// past decor, overlays and other see-through tiles: e.g. a sand shore over a
// grass edge over water. Used to decide if a see-through shore animates; a
// few stray water pixels under a sand/grass edge do not count.
func (g *Grid) WaterBelow(layer int, row, col uint32) bool {
	through := tiled.PixelMask{}
	for y := range through {
		through[y] = 0xffff
	}
	if layer >= 0 && layer < len(g.Layers) {
		if t := g.Layers[layer].TileAt(row, col); t != nil && t.SeeThrough != nil {
			through = *t.SeeThrough
		}
	}
	var below []int
	for i := 0; i < layer && i < len(g.Layers); i++ {
		if g.Layers[i].Kind == LayerKindTerrain {
			below = append(below, i)
		}
	}
	shown := exposedWater(g, below, len(below)-1, row, col, through)
	n := shown.Count()
	return n > 0 && n*2 >= through.Count()
}

type TileLayer struct {
	Name  string
	Kind  LayerKind
	Tiles map[uint32]map[uint32]*Tile
}

func NewTileLayer(name string, kind LayerKind) TileLayer {
	return TileLayer{
		Name:  name,
		Kind:  kind,
		Tiles: make(map[uint32]map[uint32]*Tile),
	}
}

func (l *TileLayer) SetTile(row, col uint32, tile *Tile) {
	if l.Tiles[row] == nil {
		l.Tiles[row] = make(map[uint32]*Tile)
	}
	l.Tiles[row][col] = tile
}

func (l *TileLayer) TileAt(row, col uint32) *Tile {
	if l.Tiles[row] == nil {
		return nil
	}
	return l.Tiles[row][col]
}

type GridTileset struct {
	Key        string
	ImagePath  string
	TileWidth  int
	TileHeight int
	Columns    int
	// Animations maps a local tile id to its frames (from the Tiled tileset).
	Animations map[int][]tiled.AnimFrame
	Synced     map[int]bool
}

type GridFileEntityInstance struct {
	Row        uint32
	Col        uint32
	Definition string
	GridID     string `yaml:"id"` // Ignored from map file for now. Overwritten by server. Will need to change this.

	// Properties are Tiled object custom properties (e.g. msg on signs).
	Properties map[string]string `yaml:"-" json:"-"`

	Spawned bool   `yaml:"-"`
	State   string `yaml:"-"`
}

// SpawnPoint marks a player spawn location loaded from a Tiled spawn tile/object.
type SpawnPoint struct {
	Row uint32
	Col uint32
	X   int32
	Y   int32
}

type GridSerializable struct {
	ColCount uint32
	RowCount uint32
	Tiles []*Tile
	Entities []GridFileEntityInstance
	LastEntityId string `yaml:"lastEntityId"`
}

type Grid struct {
	Filename string
	ColCount uint32
	RowCount uint32
	Tiles map[uint32]map[uint32]*Tile
	Layers []TileLayer
	Entities []GridFileEntityInstance
	SpawnPoints []SpawnPoint
	LastEntityId string `yaml:"lastEntityId"`
	Tilesets []GridTileset
}

func (g *Grid) BuildEmpty() {
	g.Tiles = make(map[uint32]map[uint32]*Tile)

	for row := uint32(0); row < g.RowCount; row++ {
		g.Tiles[row] = make(map[uint32]*Tile)
	}
}

// TopTileAt returns the topmost non-empty tile at a cell across all layers.
// Upper layers take precedence. In-bounds empty cells return a traversable tile.
// Out-of-bounds coordinates return nil.
// CollisionTileAt is the topmost non-decor tile: what the player stands on
// for collision, teleports and other tile rules.
func (g *Grid) CollisionTileAt(row, col uint32) *Tile {
	if row >= g.RowCount || col >= g.ColCount {
		return nil
	}
	var extra []CollisionRect
	for i := len(g.Layers) - 1; i >= 0; i-- {
		tile := g.Layers[i].TileAt(row, col)
		if tile == nil || tile.Decor {
			continue
		}
		if tile.Overlay || tile.Above {
			extra = append(extra, tile.Collision...)
			continue
		}
		// Overlays under the deciding tile (a tall wall face's lower half,
		// drawn beneath a nearer wall) still collide.
		for j := i - 1; j >= 0; j-- {
			if t := g.Layers[j].TileAt(row, col); t != nil && t.Overlay && !t.Decor {
				extra = append(extra, t.Collision...)
			}
		}
		if len(extra) == 0 {
			return tile
		}
		merged := *tile
		merged.Collision = append(append([]CollisionRect(nil), tile.Collision...), extra...)
		return &merged
	}
	top := g.TopTileAt(row, col)
	if top != nil && (top.Decor || top.Overlay || top.Above) {
		// Only decor/overlays in this cell: an empty cell plus overlay rects.
		return &Tile{Row: row, Col: col, Traversable: len(extra) == 0, Collision: extra, CollisionFromTSX: true}
	}
	return top
}

func (g *Grid) TopTileAt(row, col uint32) *Tile {
	if row >= g.RowCount || col >= g.ColCount {
		return nil
	}

	for i := len(g.Layers) - 1; i >= 0; i-- {
		if tile := g.Layers[i].TileAt(row, col); tile != nil {
			return tile
		}
	}

	if g.Tiles != nil {
		if rowTiles, ok := g.Tiles[row]; ok {
			if tile := rowTiles[col]; tile != nil {
				return tile
			}
		}
	}

	return &Tile{
		Row:         row,
		Col:         col,
		Traversable: true,
	}
}

// SyncTopTiles updates the flat Tiles map to the topmost tile per cell.
func (g *Grid) SyncTopTiles() {
	if g.Tiles == nil {
		g.BuildEmpty()
	}

	for _, pos := range g.CellIterator() {
		top := g.TopTileAt(pos.Row, pos.Col)
		if top != nil {
			g.Tiles[pos.Row][pos.Col] = top
			continue
		}

		g.Tiles[pos.Row][pos.Col] = &Tile{
			Row:         pos.Row,
			Col:         pos.Col,
			Traversable: true,
		}
	}
}

func (g *Grid) Build() {
	g.Tiles = make(map[uint32]map[uint32]*Tile)

	for row := uint32(0); row < g.RowCount; row++ {
		g.Tiles[row] = make(map[uint32]*Tile)

		for col := uint32(0); col < g.ColCount; col++ {
			t := &Tile{
				Row: row,
				Col: col,
				Texture: "water",
			}

			g.Tiles[row][col] = t
		}
	}
}

type position struct {
	Row uint32
	Col uint32
}

func (g *Grid) CellIterator() []*position {
	ret := make([]*position, 0)

	for row := uint32(0); row < g.RowCount; row++ {
		for col := uint32(0); col < g.ColCount; col++ {
			ret = append(ret, &position{
				Row: row,
				Col: col,
			})
		}
	}

	return ret
}
