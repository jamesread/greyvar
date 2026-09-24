package greyvarserver

import (
	"strings"

	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/server/pkg/worlds"
	log "github.com/sirupsen/logrus"
)

func (s *serverInterface) transitionPlayerToGrid(rp *RemotePlayer, destGridId string, x int32, y int32) bool {
	world := s.worldForPlayer(rp)
	if world == nil {
		return false
	}

	destGrid, ok := world.Grids[destGridId]
	if !ok || destGrid == nil {
		return false
	}

	oldGridId := rp.CurrentGridId
	oldWorldId := rp.CurrentWorldId

	entryX, entryY := s.resolveEntryPosition(rp.CurrentWorldId, destGridId, x, y)

	rp.PendingGridTransition = nil
	if dx, dy, ok := worlds.ScrollDeltaBetween(world, oldGridId, destGridId); ok && (dx != 0 || dy != 0) {
		rp.PendingGridTransition = &GridTransitionInfo{
			FromGridId:   oldGridId,
			ScrollDeltaX: dx,
			ScrollDeltaY: dy,
		}
	}

	for _, other := range s.remotePlayers {
		if other == rp {
			continue
		}

		if other.CurrentWorldId != oldWorldId || other.CurrentGridId != oldGridId {
			continue
		}

		if _, known := other.KnownEntities[rp.Entity.ServerId]; known {
			other.PendingDespawns = append(other.PendingDespawns, rp.Entity.ServerId)
			delete(other.KnownEntities, rp.Entity.ServerId)
		}
	}

	for id := range rp.KnownEntities {
		if id == rp.Entity.ServerId {
			continue
		}
		rp.PendingDespawns = append(rp.PendingDespawns, id)
	}

	rp.CurrentGridId = destGridId
	rp.Entity.GridId = destGridId
	rp.Entity.WorldId = rp.CurrentWorldId
	rp.Entity.X = entryX
	rp.Entity.Y = entryY
	rp.NeedsGridUpdate = true
	clearSignHud(rp)
	rp.ActiveEntityIds = make(map[int64]bool)

	log.WithFields(log.Fields{
		"player":   rp.Username,
		"entityId": rp.Entity.ServerId,
		"world":    rp.CurrentWorldId,
		"fromGrid": oldGridId,
		"toGrid":   destGridId,
		"x":        entryX,
		"y":        entryY,
	}).Info("Player transitioned grid")

	return true
}

func (s *serverInterface) resolveEntryPosition(worldId string, gridId string, x int32, y int32) (int32, int32) {
	grid := s.gridById(worldId, gridId)
	if grid == nil {
		return x, y
	}

	_, _, maxX, maxY := gridPixelBounds(grid)
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x > maxX {
		x = maxX
	}
	if y > maxY {
		y = maxY
	}

	if s.isTraversableOnGrid(worldId, gridId, x, y) {
		return x, y
	}

	// Nudge along the entry edge to find open ground (common when edge tiles have collision).
	for _, delta := range []int32{4, -4, 8, -8, 12, -12, 16, -16} {
		if s.isTraversableOnGrid(worldId, gridId, x+delta, y) {
			return x + delta, y
		}
		if s.isTraversableOnGrid(worldId, gridId, x, y+delta) {
			return x, y + delta
		}
	}

	return x, y
}

func (s *serverInterface) tryEdgeTransition(rp *RemotePlayer, mr *pb.MoveRequest, newX int32, newY int32) bool {
	grid := s.gridForPlayer(rp)
	world := s.worldForPlayer(rp)
	if grid == nil || world == nil || mr == nil {
		return false
	}

	ent := rp.Entity
	_, _, maxX, maxY := gridPixelBounds(grid)

	// Movement deltas: mr.X = col axis, mr.Y = row axis (matches WASD / arrow keys).
	if mr.Y < 0 && (ent.Y <= 0 || newY < 0) {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, -1, 0)
		if !ok {
			return false
		}

		dest := world.Grids[destId]
		_, _, _, entryY := gridPixelBounds(dest)
		return s.transitionPlayerToGrid(rp, destId, ent.X, entryY)
	}

	if mr.Y > 0 && (ent.Y >= maxY || newY > maxY) {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, 1, 0)
		if !ok {
			return false
		}

		return s.transitionPlayerToGrid(rp, destId, ent.X, 0)
	}

	if mr.X < 0 && (ent.X <= 0 || newX < 0) {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, 0, -1)
		if !ok {
			return false
		}

		dest := world.Grids[destId]
		_, _, entryX, _ := gridPixelBounds(dest)
		return s.transitionPlayerToGrid(rp, destId, entryX, ent.Y)
	}

	if mr.X > 0 && (ent.X >= maxX || newX > maxX) {
		destId, ok := worlds.AdjacentGridId(world, rp.CurrentGridId, 0, 1)
		if !ok {
			return false
		}

		return s.transitionPlayerToGrid(rp, destId, 0, ent.Y)
	}

	return false
}

func (s *serverInterface) tryTeleportTile(rp *RemotePlayer) bool {
	grid := s.gridForPlayer(rp)
	if grid == nil {
		return false
	}

	row, col := tileCoordsFromPixels(rp.Entity.X, rp.Entity.Y)
	if row >= grid.RowCount || col >= grid.ColCount {
		return false
	}

	rowTiles, ok := grid.Tiles[row]
	if !ok {
		return false
	}

	tile := rowTiles[col]
	if tile == nil || tile.TeleportDst == "" {
		return false
	}

	x := int32(tile.TeleportX * tileSizePx)
	y := int32(tile.TeleportY * tileSizePx)

	return s.transitionPlayerToGrid(rp, tile.TeleportDst, x, y)
}

func (s *serverInterface) tryTeleportPad(rp *RemotePlayer, ent *Entity, vis tiled.TypedTileVisual) {
	if rp == nil || ent == nil {
		return
	}

	worldId := strings.TrimSpace(entityProp(ent, vis, "teleport_world"))
	destination := strings.TrimSpace(entityProp(ent, vis, "teleport_destination"))

	if worldId == "" {
		s.teleportPadError(rp, ent, "Teleport failed: teleport_world is missing")
		return
	}
	if destination == "" {
		s.teleportPadError(rp, ent, "Teleport failed: teleport_destination is missing")
		return
	}

	world, err := s.ensureWorldLoaded(worldId)
	if err != nil || world == nil {
		msg := "Teleport failed: world not found (" + worldId + ")"
		if err != nil {
			msg = "Teleport failed: world not found (" + worldId + "): " + err.Error()
		}
		s.teleportPadError(rp, ent, msg)
		return
	}

	destGridId, destGrid := resolveWorldGrid(world, destination)
	if destGrid == nil {
		s.teleportPadError(rp, ent, "Teleport failed: destination not found ("+destination+")")
		return
	}

	spawnX, spawnY := spawnPositionOnGrid(destGrid)

	log.WithFields(log.Fields{
		"player":      rp.Username,
		"padEntityId": ent.ServerId,
		"world":       worldId,
		"destination": destGridId,
		"x":           spawnX,
		"y":           spawnY,
	}).Info("Teleport pad activated")

	if worldId == rp.CurrentWorldId {
		s.transitionPlayerToGrid(rp, destGridId, spawnX, spawnY)
		return
	}

	s.leaveCurrentLocation(rp)
	rp.PendingGridTransition = nil
	rp.CurrentWorldId = worldId
	rp.CurrentGridId = destGridId
	rp.Entity.WorldId = worldId
	rp.Entity.GridId = destGridId
	rp.Entity.X = spawnX
	rp.Entity.Y = spawnY
	rp.NeedsGridUpdate = true
}

func (s *serverInterface) teleportPadError(rp *RemotePlayer, ent *Entity, text string) {
	if rp == nil || text == "" {
		return
	}
	msg := text
	rp.PendingHudMessage = &msg
	if ent != nil {
		rp.PendingHudEntityId = ent.ServerId
	}
	log.WithFields(log.Fields{
		"player": rp.Username,
		"error":  text,
	}).Warn("Teleport pad error")
}

func resolveWorldGrid(world *worlds.World, destination string) (string, *gridfiles.Grid) {
	if world == nil || destination == "" {
		return "", nil
	}
	if g := world.Grids[destination]; g != nil {
		return destination, g
	}
	// Allow authors to omit the .tmj suffix.
	if !strings.HasSuffix(destination, ".tmj") {
		withExt := destination + ".tmj"
		if g := world.Grids[withExt]; g != nil {
			return withExt, g
		}
	}
	return "", nil
}

func spawnPositionOnGrid(grid *gridfiles.Grid) (int32, int32) {
	if grid == nil {
		return 0, 0
	}
	if len(grid.SpawnPoints) > 0 {
		return grid.SpawnPoints[0].X, grid.SpawnPoints[0].Y
	}
	row := grid.RowCount / 2
	col := grid.ColCount / 2
	return int32(col * tileSizePx), int32(row * tileSizePx)
}
