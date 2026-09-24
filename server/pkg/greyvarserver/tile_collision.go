package greyvarserver

import (
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/tiled"
)

func rectsOverlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}

func tileBlocksEntity(tile *gridfiles.Tile, cellRow, cellCol uint32, entityX, entityY int32) bool {
	if tile == nil {
		return true
	}

	if len(tile.Collision) > 0 {
		return entityOverlapsTileCollision(tile, cellRow, cellCol, entityX, entityY)
	}

	if tile.CollisionFromTSX {
		return false
	}

	if tile.Texture == "" {
		return false
	}

	return isBlockingTileTexture(tile.Texture)
}

func entityOverlapsTileCollision(tile *gridfiles.Tile, cellRow, cellCol uint32, entityX, entityY int32) bool {
	tileOriginX := float64(cellCol * tileSizePx)
	tileOriginY := float64(cellRow * tileSizePx)
	ex := float64(entityX)
	ey := float64(entityY)
	ew := float64(entitySizePx)
	eh := float64(entitySizePx)

	for _, colRect := range tile.Collision {
		local := tiled.TransformCollisionRect(colRect, tileSizePx, tileSizePx, tile.FlipH, tile.FlipV)
		if rectsOverlap(ex, ey, ew, eh, tileOriginX+local.X, tileOriginY+local.Y, local.W, local.H) {
			return true
		}
	}

	return false
}
