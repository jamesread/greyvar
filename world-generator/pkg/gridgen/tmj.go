package gridgen

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jamesread/greyvar/datlib/worldfiles"
)

const tileSize = 16

type tmjDoc struct {
	CompressionLevel int          `json:"compressionlevel"`
	Width            int          `json:"width"`
	Height           int          `json:"height"`
	Infinite         bool         `json:"infinite"`
	Orientation      string       `json:"orientation"`
	RenderOrder      string       `json:"renderorder"`
	TiledVersion     string       `json:"tiledversion"`
	Type             string       `json:"type"`
	Version          string       `json:"version"`
	TileWidth        int          `json:"tilewidth"`
	TileHeight       int          `json:"tileheight"`
	NextLayerID      int          `json:"nextlayerid"`
	NextObjectID     int          `json:"nextobjectid"`
	Layers           []tmjLayer   `json:"layers"`
	Tilesets         []tmjTileset `json:"tilesets"`
}

type tmjLayer struct {
	ID        int         `json:"id"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
	Width     int         `json:"width,omitempty"`
	Height    int         `json:"height,omitempty"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	Data      []uint32    `json:"data,omitempty"`
	DrawOrder string      `json:"draworder,omitempty"`
	Objects   []tmjObject `json:"objects,omitempty"`
	Opacity   float64     `json:"opacity"`
	Visible   bool        `json:"visible"`
}

type tmjObject struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation float64 `json:"rotation"`
	Visible  bool    `json:"visible"`
}

type tmjTileset struct {
	FirstGID int    `json:"firstgid"`
	Source   string `json:"source"`
}

type tiledProperty struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type worldDoc struct {
	Type                 string                    `json:"type"`
	Maps                 []worldfiles.MapPlacement `json:"maps"`
	OnlyShowAdjacentMaps bool                      `json:"onlyShowAdjacentMaps"`
	Properties           []tiledProperty           `json:"properties"`
}

func writeJSON(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}
