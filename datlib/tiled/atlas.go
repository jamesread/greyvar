package tiled

import (
	"path/filepath"
	"sort"
	"strings"
)

// AtlasTileset describes a Tiled TSX tileset backed by a single spritesheet image.
type AtlasTileset struct {
	Key        string
	ImagePath  string
	TileWidth  int
	TileHeight int
	Columns    int
	// Animations maps a local tile id to its Tiled <animation> frames.
	Animations map[int][]AnimFrame
	// Synced animations (tile property synced=true) run in step in every
	// cell, so patterns flow across tiles (water); others start per cell.
	Synced map[int]bool
}

// DefaultAnimFrameMs is used for animation frames with no (or zero) duration.
const DefaultAnimFrameMs = 100

// SanitizeAnimations drops frames that reference tiles outside the tileset
// (tileCount <= 0 disables that check), defaults missing durations, and drops
// animations left with fewer than two frames (nothing to animate).
func SanitizeAnimations(anims map[int][]AnimFrame, tileCount int) map[int][]AnimFrame {
	out := map[int][]AnimFrame{}
	for id, frames := range anims {
		var kept []AnimFrame
		for _, f := range frames {
			if f.TileID < 0 || (tileCount > 0 && f.TileID >= tileCount) {
				continue
			}
			if f.DurationMs <= 0 {
				f.DurationMs = DefaultAnimFrameMs
			}
			kept = append(kept, f)
		}
		if len(kept) >= 2 {
			out[id] = kept
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ResWebPath converts a filesystem path that contains the project's res/ tree
// into a URL path relative to the web client's /res/ base
// (e.g. img/textures/tilesets/grass.png).
func ResWebPath(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(path))
	idx := strings.Index(normalized, "res/")
	if idx < 0 {
		return ""
	}

	suffix := normalized[idx+len("res/"):]
	for _, prefix := range []string{"img/", "atlas/", "entdefs/"} {
		if strings.HasPrefix(suffix, prefix) {
			return suffix
		}
	}

	return ""
}

func (c *TilesetCatalog) registerAtlas(atlas AtlasTileset) {
	if atlas.Key == "" || atlas.ImagePath == "" {
		return
	}
	if c.atlases == nil {
		c.atlases = make(map[string]AtlasTileset)
	}
	if _, exists := c.atlases[atlas.Key]; !exists {
		c.atlases[atlas.Key] = atlas
	}
}

// AtlasTilesets returns atlas tilesets referenced by the loaded map, sorted by key.
func (c *TilesetCatalog) AtlasTilesets() []AtlasTileset {
	if len(c.atlases) == 0 {
		return nil
	}

	result := make([]AtlasTileset, 0, len(c.atlases))
	for _, atlas := range c.atlases {
		result = append(result, atlas)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result
}
