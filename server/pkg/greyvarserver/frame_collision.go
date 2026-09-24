package greyvarserver

import (
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/tiled"
)

const tileSizePx = 16
const entitySizePx = 16

func isTileTraversable(tile *gridfiles.Tile) bool {
	if tile == nil {
		return false
	}

	if len(tile.Collision) > 0 {
		return false
	}

	if tile.CollisionFromTSX {
		return true
	}

	if tile.Texture == "" {
		return true
	}

	if isBlockingTileTexture(tile.Texture) {
		return false
	}

	return true
}

func tileCoordsFromPixels(x int32, y int32) (uint32, uint32) {
	// Grid yaml row/col match screen space: x = col, y = row (see client GridScene).
	row := uint32(y / tileSizePx)
	col := uint32(x / tileSizePx)

	return row, col
}

func gridPixelBounds(grid *gridfiles.Grid) (widthPx int32, heightPx int32, maxEntityX int32, maxEntityY int32) {
	if grid == nil {
		return 0, 0, 0, 0
	}

	widthPx = int32(grid.ColCount) * tileSizePx
	heightPx = int32(grid.RowCount) * tileSizePx
	maxEntityX = widthPx - entitySizePx
	maxEntityY = heightPx - entitySizePx
	return widthPx, heightPx, maxEntityX, maxEntityY
}

func entityWithinGridBounds(grid *gridfiles.Grid, x int32, y int32) bool {
	_, _, maxX, maxY := gridPixelBounds(grid)
	return x >= 0 && y >= 0 && x <= maxX && y <= maxY
}

func (s *serverInterface) gridForPlayer(rp *RemotePlayer) *gridfiles.Grid {
	if rp == nil {
		return nil
	}

	return s.gridById(rp.CurrentWorldId, rp.CurrentGridId)
}

func (s *serverInterface) tileAtGrid(grid *gridfiles.Grid, x int32, y int32) *gridfiles.Tile {
	if grid == nil {
		return nil
	}

	row, col := tileCoordsFromPixels(x, y)

	if row >= grid.RowCount || col >= grid.ColCount {
		return nil
	}

	return grid.TopTileAt(row, col)
}

func (s *serverInterface) tileAtForPlayer(rp *RemotePlayer, x int32, y int32) *gridfiles.Tile {
	return s.tileAtGrid(s.gridForPlayer(rp), x, y)
}

func (s *serverInterface) tileAtGridCell(grid *gridfiles.Grid, row uint32, col uint32) *gridfiles.Tile {
	if grid == nil {
		return nil
	}

	if row >= grid.RowCount || col >= grid.ColCount {
		return nil
	}

	return grid.TopTileAt(row, col)
}

func (s *serverInterface) isTraversableOnGrid(worldId string, gridId string, x int32, y int32) bool {
	grid := s.gridById(worldId, gridId)
	if grid == nil {
		return false
	}

	if !entityWithinGridBounds(grid, x, y) {
		return false
	}

	minRow := uint32(y / tileSizePx)
	maxRow := uint32((y + entitySizePx - 1) / tileSizePx)
	minCol := uint32(x / tileSizePx)
	maxCol := uint32((x + entitySizePx - 1) / tileSizePx)

	for row := minRow; row <= maxRow; row++ {
		for col := minCol; col <= maxCol; col++ {
			tile := s.tileAtGridCell(grid, row, col)
			if tileBlocksEntity(tile, row, col, x, y) {
				return false
			}
		}
	}

	return true
}

func (s *serverInterface) isTraversableAt(rp *RemotePlayer, x int32, y int32) bool {
	if rp == nil {
		return false
	}

	return s.isTraversableOnGrid(rp.CurrentWorldId, rp.CurrentGridId, x, y)
}

func (s *serverInterface) isTraversableForPlayer(rp *RemotePlayer, x int32, y int32) bool {
	return s.isTraversableAt(rp, x, y)
}

func (s *serverInterface) isBlockedByPlayer(rp *RemotePlayer, ent *Entity, newX int32, newY int32) bool {
	for _, other := range s.entityInstances {
		if other.ServerId == ent.ServerId || other.Definition != "player" {
			continue
		}

		if other.GridId != rp.CurrentGridId || other.WorldId != rp.CurrentWorldId {
			continue
		}

		if other.X == newX && other.Y == newY {
			return true
		}
	}

	return false
}

func (s *serverInterface) isBlockedBySolidEntity(rp *RemotePlayer, mover *Entity, newX int32, newY int32) (bool, string) {
	if rp == nil || mover == nil {
		return false, ""
	}

	for _, other := range s.entitiesOnGrid(rp.CurrentWorldId, rp.CurrentGridId) {
		if other.ServerId == mover.ServerId {
			continue
		}

		entdef := s.entityDefinitions[other.Definition]
		if entdef == nil {
			continue
		}

		title := entdef.Title
		if title == "" {
			title = other.Definition
		}

		vis, hasVis := s.entityVisuals.ResolveState(title, other.State, entdef.InitialState)
		ct := tiled.CollisionTouch
		if hasVis {
			ct = vis.CollisionType.OrDefault()
		}
		hasShapes := hasVis && len(vis.Collision) > 0

		// pressure / pickup are walk-through; only touch blocks.
		if !ct.BlocksMovement() {
			continue
		}

		if !entdef.Solid && !hasShapes {
			continue
		}

		if hasShapes {
			if entityOverlapsEntityCollision(vis.Collision, other.X, other.Y, newX, newY) {
				return true, other.Definition
			}
			continue
		}

		if rectsOverlap(
			float64(newX), float64(newY), float64(entitySizePx), float64(entitySizePx),
			float64(other.X), float64(other.Y), float64(entitySizePx), float64(entitySizePx),
		) {
			return true, other.Definition
		}
	}

	return false, ""
}

func entityOverlapsEntityCollision(rects []tiled.CollisionRect, otherX, otherY, moverX, moverY int32) bool {
	moverL := float64(moverX)
	moverT := float64(moverY)
	moverR := moverL + float64(entitySizePx)
	moverB := moverT + float64(entitySizePx)

	for _, rect := range rects {
		l := float64(otherX) + rect.X
		t := float64(otherY) + rect.Y
		r := l + rect.W
		b := t + rect.H
		if moverL < r && moverR > l && moverT < b && moverB > t {
			return true
		}
	}
	return false
}

func (s *serverInterface) moveBlockReason(rp *RemotePlayer, ent *Entity, newX int32, newY int32) string {
	if !s.isTraversableAt(rp, newX, newY) {
		grid := s.gridForPlayer(rp)
		texture := ""
		if grid != nil {
			row, col := tileCoordsFromPixels(newX, newY)
			if tile := s.tileAtGridCell(grid, row, col); tile != nil {
				texture = tile.Texture
			}
		}
		return "tile blocked (" + texture + ")"
	}

	if s.isBlockedByPlayer(rp, ent, newX, newY) {
		return "player at destination"
	}

	if blocked, name := s.isBlockedBySolidEntity(rp, ent, newX, newY); blocked {
		return "entity blocked (" + name + ")"
	}

	return ""
}

func (s *serverInterface) canMoveTo(rp *RemotePlayer, ent *Entity, newX int32, newY int32) bool {
	return s.moveBlockReason(rp, ent, newX, newY) == ""
}

func (s *serverInterface) entitiesOnGrid(worldId string, gridId string) []*Entity {
	out := make([]*Entity, 0)

	for _, ent := range s.entityInstances {
		if ent.GridId == gridId && ent.WorldId == worldId {
			out = append(out, ent)
		}
	}

	return out
}
