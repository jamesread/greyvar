package greyvarserver

import (
	"strconv"
	"strings"

	"github.com/jamesread/greyvar/datlib/tiled"
	log "github.com/sirupsen/logrus"
)

// Puzzle mechanics that move or time out:
//
//	push blocks   entdef `pushable: true` (and solid): keep walking into one
//	              and it slides a tile, if that tile is free floor; the use
//	              button pulls it a tile towards you (you step back)
//	plates        pressure entities (not teleporters) stay pressed while any
//	              player or push block is on them
//	reset_after   map property (ms) on a toggle entity (a lever or wallSwitch): once
//	              flipped away from its initial state it springs back

// pushHoldMoves is how many blocked moves in a row (at the client's move
// rate, about 15 a second) push a block: a short lean, not a brush past.
const pushHoldMoves = 4

// pushAlignPx: the player's hitbox must overlap the block this much across
// the push direction (so you push it square on, not by its corner).
const pushAlignPx = 6

// tryPush is called for a blocked cardinal move: when the player has been
// leaning on a pushable entity long enough, slide it one tile.
func (s *serverInterface) tryPush(rp *RemotePlayer, dx, dy int32) bool {
	if rp == nil || rp.Entity == nil || (dx != 0) == (dy != 0) {
		s.resetPush(rp)
		return false
	}
	block := s.pushableAhead(rp, dx, dy)
	if block == nil {
		s.resetPush(rp)
		return false
	}
	if rp.pushTarget != block.ServerId || rp.pushDirX != dx || rp.pushDirY != dy {
		rp.pushTarget, rp.pushDirX, rp.pushDirY, rp.pushCount = block.ServerId, dx, dy, 0
	}
	rp.pushCount++
	if rp.pushCount < pushHoldMoves {
		return false
	}
	rp.pushCount = 0
	return s.pushBlock(block, dx, dy)
}

func (s *serverInterface) resetPush(rp *RemotePlayer) {
	if rp != nil {
		rp.pushTarget, rp.pushCount = 0, 0
	}
}

// pushableAhead is the pushable entity the player's next pixel in (dx, dy)
// would run into, if the player is square on to it.
func (s *serverInterface) pushableAhead(rp *RemotePlayer, dx, dy int32) *Entity {
	pl, pt, pw, ph := moverBox(rp.Entity.X+dx, rp.Entity.Y+dy)
	for _, ent := range s.entitiesOnGrid(rp.CurrentWorldId, rp.CurrentGridId) {
		entdef := s.entityDefinitions[ent.Definition]
		if entdef == nil || !entdef.Pushable {
			continue
		}
		bx, by := float64(ent.X), float64(ent.Y)
		if !rectsOverlap(pl, pt, pw, ph, bx, by, entitySizePx, entitySizePx) {
			continue
		}
		var across float64
		if dx != 0 {
			across = min(pt+ph, by+entitySizePx) - max(pt, by)
		} else {
			across = min(pl+pw, bx+entitySizePx) - max(pl, bx)
		}
		if across >= pushAlignPx {
			return ent
		}
	}
	return nil
}

// pushBlock slides block one tile in (dx, dy) if it fits there.
func (s *serverInterface) pushBlock(block *Entity, dx, dy int32) bool {
	nx, ny := block.X+dx*tileSizePx, block.Y+dy*tileSizePx
	if !s.blockCanEnter(block, nx, ny) {
		return false
	}
	block.X, block.Y = nx, ny
	log.Debugf("pushed %s %d to %d,%d", block.Definition, block.ServerId, nx, ny)
	s.refreshPlates(block.WorldId, block.GridId)
	return true
}

// blockCanEnter: the tile cell at (x, y) is open floor with nothing solid,
// no door, no teleporter and no player in it.
func (s *serverInterface) blockCanEnter(block *Entity, x, y int32) bool {
	grid := s.gridById(block.WorldId, block.GridId)
	if grid == nil || !entityWithinGridBounds(grid, x, y) || x%tileSizePx != 0 || y%tileSizePx != 0 {
		return false
	}
	row, col := tileCoordsFromPixels(x, y)
	if !isTileTraversable(s.tileAtGridCell(grid, row, col)) {
		return false
	}
	for _, other := range s.entitiesOnGrid(block.WorldId, block.GridId) {
		if other.ServerId == block.ServerId {
			continue
		}
		if other.Definition == "player" {
			pl, pt, pw, ph := moverBox(other.X, other.Y)
			if rectsOverlap(pl, pt, pw, ph, float64(x), float64(y), entitySizePx, entitySizePx) {
				return false
			}
			continue
		}
		if absInt32(other.X-x) >= entitySizePx || absInt32(other.Y-y) >= entitySizePx {
			continue
		}
		entdef := s.entityDefinitions[other.Definition]
		if entdef == nil {
			continue
		}
		vis, ct, ok := s.entityVisualInfo(other)
		switch {
		case entdef.Solid, len(entdef.PassableStates) > 0: // walls, doors (even open)
			return false
		case ok && ct == tiled.CollisionPressure && isTeleporter(other, vis):
			return false
		case ok && ct.BlocksMovement() && len(vis.Collision) > 0:
			return false
		}
	}
	return true
}

func isTeleporter(ent *Entity, vis tiled.TypedTileVisual) bool {
	return ent.Definition == "teleportPad" || isRiftPortal(ent) || entityProp(ent, vis, "teleport_destination") != ""
}

// pressTransition is the state a pressure entity takes while held down.
func pressTransition(ent *Entity, vis tiled.TypedTileVisual) string {
	if t := entityProp(ent, vis, "collision_transition"); t != "" {
		return t
	}
	return "pressed"
}

// plateHeld: a player stands mostly on the plate, or a push block sits on it.
// skip is an entity to ignore (a player who is leaving the grid).
func (s *serverInterface) plateHeld(plate *Entity, vis tiled.TypedTileVisual, skip int64) bool {
	rects := triggerRectsForVisual(vis)
	if len(rects) == 0 {
		return false
	}
	for _, other := range s.entitiesOnGrid(plate.WorldId, plate.GridId) {
		if other.ServerId == plate.ServerId || other.ServerId == skip {
			continue
		}
		if other.Definition == "player" {
			if playerMajorityOverlaps(other.X, other.Y, rects, plate.X, plate.Y) {
				return true
			}
			continue
		}
		entdef := s.entityDefinitions[other.Definition]
		if entdef == nil || !entdef.Pushable {
			continue
		}
		var plateArea, covered float64
		for _, r := range rects {
			plateArea += r.W * r.H
			covered += aabbIntersectionArea(float64(other.X), float64(other.Y), entitySizePx, entitySizePx,
				float64(plate.X)+r.X, float64(plate.Y)+r.Y, r.W, r.H)
		}
		if covered*2 > plateArea {
			return true
		}
	}
	return false
}

// refreshPlate presses or releases one pressure plate by who is on it.
func (s *serverInterface) refreshPlate(plate *Entity) { s.refreshPlateSkipping(plate, 0) }

func (s *serverInterface) refreshPlateSkipping(plate *Entity, skip int64) {
	vis, ct, ok := s.entityVisualInfo(plate)
	if !ok || ct != tiled.CollisionPressure || isTeleporter(plate, vis) {
		return
	}
	pressed := pressTransition(plate, vis)
	if s.plateHeld(plate, vis, skip) {
		s.setEntityState(plate, pressed)
		return
	}
	if plate.State != pressed {
		return
	}
	initial := "unpressed"
	if entdef := s.entityDefinitions[plate.Definition]; entdef != nil && entdef.InitialState != "" {
		initial = entdef.InitialState
	}
	s.setEntityState(plate, initial)
}

// refreshPlates re-checks every pressure plate on a grid, ignoring the
// entity skip (0 for none).
func (s *serverInterface) refreshPlates(worldId, gridId string, skip ...int64) {
	var sk int64
	if len(skip) > 0 {
		sk = skip[0]
	}
	for _, ent := range s.entitiesOnGrid(worldId, gridId) {
		if ent.Definition != "player" {
			s.refreshPlateSkipping(ent, sk)
		}
	}
}

// resetAfterMs is a toggle entity's spring-back time (0 = stays put).
func resetAfterMs(ent *Entity) int64 {
	ms, err := strconv.ParseInt(strings.TrimSpace(ent.Properties["reset_after"]), 10, 64)
	if err != nil || ms <= 0 {
		return 0
	}
	return ms
}

// armReset starts (or stops) a reset_after countdown after ent was toggled
// by rp. Returns whether it is now counting down.
func (s *serverInterface) armReset(rp *RemotePlayer, ent *Entity, initial string) bool {
	ms := resetAfterMs(ent)
	if ms == 0 {
		return false
	}
	if ent.State == initial {
		ent.resetAt = 0
		return false
	}
	ent.resetAt = nowMs() + ms
	if rp != nil {
		sayAt(rp, ent, "Tick, tock... hurry!")
	}
	return true
}

// tickTimers springs back every lever whose reset_after has run out.
func (s *serverInterface) tickTimers(now int64) {
	for _, ent := range s.entityInstances {
		if ent.resetAt == 0 || now < ent.resetAt {
			continue
		}
		ent.resetAt = 0
		if entdef := s.entityDefinitions[ent.Definition]; entdef != nil && entdef.InitialState != "" {
			s.setEntityState(ent, entdef.InitialState)
		}
	}
}

// pullReachPx: how far the player's hitbox may be from the block's side to
// pull it (standing in the next cell is enough).
const pullReachPx = 8

// pullBlock: the use button on a pushable block the player is square on to
// (beside it, not diagonal) draws it one tile towards them while they step
// back a tile, so a block shoved against a wall or into a corner can be
// recovered. Both must fit: the player's new spot and the block's.
func (s *serverInterface) pullBlock(rp *RemotePlayer, block *Entity) bool {
	if rp == nil || rp.Entity == nil || block == nil {
		return false
	}
	pl, pt, pw, ph := moverBox(rp.Entity.X, rp.Entity.Y)
	bx, by := float64(block.X), float64(block.Y)
	// Which side of the block the player is on.
	cx, cy := pl+pw/2-(bx+entitySizePx/2), pt+ph/2-(by+entitySizePx/2)
	var dx, dy int32
	var gap, across float64
	if absf(cx) >= absf(cy) {
		dx = int32(signf(cx))
		gap = max(pl-(bx+entitySizePx), bx-(pl+pw))
		across = min(pt+ph, by+entitySizePx) - max(pt, by)
	} else {
		dy = int32(signf(cy))
		gap = max(pt-(by+entitySizePx), by-(pt+ph))
		across = min(pl+pw, bx+entitySizePx) - max(pl, bx)
	}
	if gap > pullReachPx || across < pushAlignPx {
		return false
	}
	ox, oy := rp.Entity.X, rp.Entity.Y
	nx, ny := ox+dx*tileSizePx, oy+dy*tileSizePx
	if !s.canMoveTo(rp, rp.Entity, nx, ny) {
		return false
	}
	rp.Entity.X, rp.Entity.Y = nx, ny
	if !s.pushBlock(block, dx, dy) {
		rp.Entity.X, rp.Entity.Y = ox, oy
		return false
	}
	s.resetPush(rp)
	s.processEntityInteractions(rp)
	return true
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func signf(v float64) float64 {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}
