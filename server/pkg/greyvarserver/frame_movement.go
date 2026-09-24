package greyvarserver

import (
	"math"

	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

const DISTANCE_PER_REQUEST = 6

// diagonalStep is ≈ DISTANCE_PER_REQUEST / √2 so diagonal speed matches cardinal.
const diagonalStep = 4

// entityTriggerMarginPx: player AABB must come this close to touch-collision
// shapes to trigger. Players are blocked by those shapes, so triggers use
// proximity rather than overlap.
const entityTriggerMarginPx = 3

// entityTriggerOriginPx: fallback chebyshev proximity for touch entities with
// no tileset collision shapes.
const entityTriggerOriginPx = 12

func processMoveRequest(s *serverInterface, rp *RemotePlayer, mr *pb.MoveRequest) {
	fields := log.Fields{
		"player":   rp.Username,
		"entityId": rp.Entity.ServerId,
		"gridId":   rp.CurrentGridId,
		"fromX":    rp.Entity.X,
		"fromY":    rp.Entity.Y,
		"deltaX":   mr.X,
		"deltaY":   mr.Y,
	}

	if (s.frameTime - rp.TimeOfLastMoveRequest) < 67 {
		log.WithFields(fields).Info("MoveRequest throttled")
		return
	}

	if mr.X < -1 || mr.X > 1 || mr.Y < -1 || mr.Y > 1 {
		log.WithFields(fields).Info("MoveRequest rejected: delta out of range")
		return
	}

	if mr.X == 0 && mr.Y == 0 {
		return
	}

	step := int32(DISTANCE_PER_REQUEST)
	if mr.X != 0 && mr.Y != 0 {
		step = diagonalStep
	}

	newX := rp.Entity.X + (mr.X * step)
	newY := rp.Entity.Y + (mr.Y * step)
	fields["toX"] = newX
	fields["toY"] = newY

	if s.tryEdgeTransition(rp, mr, newX, newY) {
		rp.TimeOfLastMoveRequest = s.frameTime
		log.WithFields(fields).Info("MoveRequest triggered grid transition")
		s.processEntityInteractions(rp)
		return
	}

	destX, destY, ok := s.resolveMoveDestination(rp, newX, newY)
	if !ok {
		log.WithFields(fields).Info("MoveRequest blocked")
		rp.TimeOfLastMoveRequest = s.frameTime
		s.processEntityInteractions(rp)
		return
	}

	rp.Entity.X = destX
	rp.Entity.Y = destY
	rp.TimeOfLastMoveRequest = s.frameTime
	fields["toX"] = destX
	fields["toY"] = destY

	log.WithFields(fields).Info("MoveRequest accepted")

	s.processEntityInteractions(rp)

	if s.tryTeleportTile(rp) {
		log.WithFields(fields).Info("MoveRequest triggered teleport tile")
	}
}

// resolveMoveDestination prefers the full step. If blocked, walks as far as
// possible along the intended vector so the player AABB ends flush with
// colliders (0px gap). On diagonal blocks, falls back to per-axis slides.
func (s *serverInterface) resolveMoveDestination(rp *RemotePlayer, newX, newY int32) (int32, int32, bool) {
	fromX := rp.Entity.X
	fromY := rp.Entity.Y
	dx := newX - fromX
	dy := newY - fromY

	if destX, destY, ok := s.farthestWalkableAlong(rp, fromX, fromY, dx, dy); ok {
		return destX, destY, true
	}

	if dx != 0 && dy != 0 {
		if destX, destY, ok := s.farthestWalkableAlong(rp, fromX, fromY, dx, 0); ok {
			return destX, destY, true
		}
		if destX, destY, ok := s.farthestWalkableAlong(rp, fromX, fromY, 0, dy); ok {
			return destX, destY, true
		}
	}

	return fromX, fromY, false
}

// farthestWalkableAlong returns the farthest position from (fromX,fromY) toward
// (fromX+dx, fromY+dy) that is still walkable. For diagonals, dx/dy shrink
// together so movement stays on the intended diagonal.
func (s *serverInterface) farthestWalkableAlong(rp *RemotePlayer, fromX, fromY, dx, dy int32) (int32, int32, bool) {
	if dx == 0 && dy == 0 {
		return fromX, fromY, false
	}

	absX := absInt32(dx)
	absY := absInt32(dy)
	steps := absX
	if absY > steps {
		steps = absY
	}
	if absX != 0 && absY != 0 {
		// Keep diagonal moves on the intended 45° line.
		steps = absX
		if absY < steps {
			steps = absY
		}
	}

	signX := int32(0)
	signY := int32(0)
	if dx != 0 {
		signX = dx / absX
	}
	if dy != 0 {
		signY = dy / absY
	}

	// Diagonal: advance both axes by k. Cardinal: advance the non-zero axis by k.
	for k := steps; k >= 1; k-- {
		tx := fromX + signX*k
		ty := fromY + signY*k
		if s.moveBlockReason(rp, rp.Entity, tx, ty) == "" {
			return tx, ty, true
		}
	}

	return fromX, fromY, false
}

func absInt32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

type entityContact struct {
	ent  *Entity
	vis  tiled.TypedTileVisual
	ct   tiled.CollisionType
	dist float64
}

// processEntityInteractions evaluates touch / pressure / pickup contacts after
// movement (or when blocked against a touch entity).
func (s *serverInterface) processEntityInteractions(rp *RemotePlayer) {
	if rp == nil || rp.Entity == nil {
		return
	}
	if rp.ActiveEntityIds == nil {
		rp.ActiveEntityIds = make(map[int64]bool)
	}

	current := make(map[int64]entityContact)
	for _, ent := range s.entitiesOnGrid(rp.CurrentWorldId, rp.CurrentGridId) {
		if ent.Definition == "player" || ent.ServerId == rp.Entity.ServerId {
			continue
		}
		vis, ct, ok := s.entityVisualInfo(ent)
		if !ok {
			continue
		}
		triggered, dist := s.entityContactTriggered(rp.Entity, ent, vis, ct)
		if !triggered {
			continue
		}
		current[ent.ServerId] = entityContact{ent: ent, vis: vis, ct: ct, dist: dist}
	}

	// Pickups: majority overlap → inventory + despawn (handle before HUD/state).
	for id, contact := range current {
		if contact.ct != tiled.CollisionPickup {
			continue
		}
		if rp.ActiveEntityIds[id] {
			continue // already processed
		}
		s.pickupEntity(rp, contact.ent)
		delete(current, id)
	}

	// Leave events.
	for id := range rp.ActiveEntityIds {
		if _, still := current[id]; still {
			continue
		}
		ent := s.entityInstances[id]
		if ent != nil {
			s.onEntityContactLeave(rp, ent)
		}
		delete(rp.ActiveEntityIds, id)
	}

	// Enter / stay events.
	var nearestTouchMsg *entityContact
	skipTouchHud := false
	for id, contact := range current {
		entered := !rp.ActiveEntityIds[id]
		rp.ActiveEntityIds[id] = true

		if contact.ct == tiled.CollisionTouch {
			msg := entityProp(contact.ent, contact.vis, "msg")
			if msg != "" {
				if nearestTouchMsg == nil || contact.dist < nearestTouchMsg.dist {
					c := contact
					nearestTouchMsg = &c
				}
			} else if entered {
				s.activateEntity(rp, contact.ent, contact.vis)
			}
			continue
		}

		if contact.ct == tiled.CollisionPressure && entered {
			if contact.ent.Definition == "teleportPad" {
				s.tryTeleportPad(rp, contact.ent, contact.vis)
				// Keep teleport success/error HUD; updateTouchHud would clear it.
				skipTouchHud = true
				continue
			}
			s.activateEntity(rp, contact.ent, contact.vis)
		}
	}

	if !skipTouchHud {
		s.updateTouchHud(rp, nearestTouchMsg)
	}
}

func (s *serverInterface) entityContactTriggered(player, ent *Entity, vis tiled.TypedTileVisual, ct tiled.CollisionType) (bool, float64) {
	rects := triggerRectsForVisual(vis)
	switch ct {
	case tiled.CollisionPressure, tiled.CollisionPickup:
		if len(rects) == 0 {
			return false, math.Inf(1)
		}
		ok := playerMajorityOverlaps(player.X, player.Y, rects, ent.X, ent.Y)
		dist := aabbDistanceToCollisionRects(
			float64(player.X), float64(player.Y), float64(entitySizePx), float64(entitySizePx),
			rects, ent.X, ent.Y,
		)
		return ok, dist
	default: // touch
		if len(rects) > 0 {
			dist := aabbDistanceToCollisionRects(
				float64(player.X), float64(player.Y), float64(entitySizePx), float64(entitySizePx),
				rects, ent.X, ent.Y,
			)
			return dist <= float64(entityTriggerMarginPx), dist
		}
		dX := math.Abs(float64(player.X - ent.X))
		dY := math.Abs(float64(player.Y - ent.Y))
		chebyshev := math.Max(dX, dY)
		return chebyshev < float64(entityTriggerOriginPx), chebyshev
	}
}

func triggerRectsForVisual(vis tiled.TypedTileVisual) []tiled.CollisionRect {
	if len(vis.Collision) > 0 {
		return vis.Collision
	}
	if vis.CollisionType.OrDefault() == tiled.CollisionPressure || vis.CollisionType.OrDefault() == tiled.CollisionPickup {
		w := float64(vis.TileWidth)
		h := float64(vis.TileHeight)
		if w <= 0 {
			w = float64(entitySizePx)
		}
		if h <= 0 {
			h = float64(entitySizePx)
		}
		return []tiled.CollisionRect{{X: 0, Y: 0, W: w, H: h}}
	}
	return nil
}

func playerMajorityOverlaps(playerX, playerY int32, rects []tiled.CollisionRect, originX, originY int32) bool {
	playerArea := float64(entitySizePx * entitySizePx)
	if playerArea <= 0 {
		return false
	}
	var overlap float64
	px, py := float64(playerX), float64(playerY)
	pw, ph := float64(entitySizePx), float64(entitySizePx)
	for _, rect := range rects {
		bx := float64(originX) + rect.X
		by := float64(originY) + rect.Y
		overlap += aabbIntersectionArea(px, py, pw, ph, bx, by, rect.W, rect.H)
	}
	return overlap > playerArea*0.5
}

func aabbIntersectionArea(ax, ay, aw, ah, bx, by, bw, bh float64) float64 {
	left := math.Max(ax, bx)
	right := math.Min(ax+aw, bx+bw)
	top := math.Max(ay, by)
	bottom := math.Min(ay+ah, by+bh)
	if right <= left || bottom <= top {
		return 0
	}
	return (right - left) * (bottom - top)
}

func (s *serverInterface) entityVisualInfo(ent *Entity) (tiled.TypedTileVisual, tiled.CollisionType, bool) {
	if s == nil || ent == nil {
		return tiled.TypedTileVisual{}, tiled.CollisionTouch, false
	}
	entdef := s.entityDefinitions[ent.Definition]
	title := ent.Definition
	initial := ""
	if entdef != nil {
		if entdef.Title != "" {
			title = entdef.Title
		}
		initial = entdef.InitialState
	}
	vis, ok := s.entityVisuals.ResolveState(title, ent.State, initial)
	if !ok {
		return tiled.TypedTileVisual{}, tiled.CollisionTouch, false
	}
	ct := vis.CollisionType.OrDefault()
	return vis, ct, true
}

func entityProp(ent *Entity, vis tiled.TypedTileVisual, key string) string {
	if ent != nil && ent.Properties != nil {
		if v := ent.Properties[key]; v != "" {
			return v
		}
	}
	if vis.TileProps != nil {
		return vis.TileProps[key]
	}
	return ""
}

func (s *serverInterface) activateEntity(rp *RemotePlayer, ent *Entity, vis tiled.TypedTileVisual) {
	log.Infof("Entity activated %s id=%d by %s", ent.Definition, ent.ServerId, rp.Username)

	next := entityProp(ent, vis, "collision_transition")
	if next == "" {
		next = "pressed"
	}
	if ent.State == next {
		return
	}
	ent.State = next
	rp.currentFrame.EntityStateChanges = append(rp.currentFrame.EntityStateChanges, &pb.EntityStateChange{
		EntityId: ent.ServerId,
		NewState: ent.State,
	})
}

func (s *serverInterface) onEntityContactLeave(rp *RemotePlayer, ent *Entity) {
	vis, ct, ok := s.entityVisualInfo(ent)
	if !ok || ct != tiled.CollisionPressure {
		return
	}
	entdef := s.entityDefinitions[ent.Definition]
	initial := "unpressed"
	if entdef != nil && entdef.InitialState != "" {
		initial = entdef.InitialState
	}
	transition := entityProp(ent, vis, "collision_transition")
	if transition == "" {
		transition = "pressed"
	}
	if ent.State != transition || ent.State == initial {
		return
	}
	ent.State = initial
	rp.currentFrame.EntityStateChanges = append(rp.currentFrame.EntityStateChanges, &pb.EntityStateChange{
		EntityId: ent.ServerId,
		NewState: ent.State,
	})
}

func (s *serverInterface) pickupEntity(rp *RemotePlayer, ent *Entity) {
	log.Infof("Pickup %s id=%d by %s", ent.Definition, ent.ServerId, rp.Username)

	rp.addInventoryItem(ent.Definition, 1)
	rp.PendingInventoryUpdate = true

	label := ent.Definition
	text := "Picked up " + label
	rp.PendingHudMessage = &text
	rp.PendingHudEntityId = ent.ServerId

	// Despawn for all players who know this entity.
	delete(s.entityInstances, ent.ServerId)
	for _, other := range s.remotePlayers {
		if _, known := other.KnownEntities[ent.ServerId]; known {
			other.PendingDespawns = append(other.PendingDespawns, ent.ServerId)
			delete(other.KnownEntities, ent.ServerId)
		}
		if other.ActiveEntityIds != nil {
			delete(other.ActiveEntityIds, ent.ServerId)
		}
	}
}

func (s *serverInterface) updateTouchHud(rp *RemotePlayer, nearest *entityContact) {
	msg := ""
	id := int64(0)
	if nearest != nil {
		id = nearest.ent.ServerId
		msg = entityProp(nearest.ent, nearest.vis, "msg")
	}

	if id == rp.ActiveSignEntityId && msg == rp.ActiveSignMsg {
		return
	}

	rp.ActiveSignEntityId = id
	rp.ActiveSignMsg = msg
	text := msg
	rp.PendingHudMessage = &text
	rp.PendingHudEntityId = id
}

func clearSignHud(rp *RemotePlayer) {
	if rp == nil {
		return
	}
	if rp.ActiveSignEntityId == 0 && rp.ActiveSignMsg == "" {
		return
	}
	rp.ActiveSignEntityId = 0
	rp.ActiveSignMsg = ""
	empty := ""
	rp.PendingHudMessage = &empty
	rp.PendingHudEntityId = 0
	if rp.ActiveEntityIds != nil {
		clear(rp.ActiveEntityIds)
	}
}

func aabbDistanceToCollisionRects(ax, ay, aw, ah float64, rects []tiled.CollisionRect, originX, originY int32) float64 {
	best := math.Inf(1)
	for _, rect := range rects {
		bx := float64(originX) + rect.X
		by := float64(originY) + rect.Y
		dist := aabbSeparation(ax, ay, aw, ah, bx, by, rect.W, rect.H)
		if dist < best {
			best = dist
		}
	}
	return best
}

func aabbSeparation(ax, ay, aw, ah, bx, by, bw, bh float64) float64 {
	dx := math.Max(ax-(bx+bw), bx-(ax+aw))
	dy := math.Max(ay-(by+bh), by-(ay+ah))
	if dx < 0 && dy < 0 {
		return 0
	}
	if dx < 0 {
		return dy
	}
	if dy < 0 {
		return dx
	}
	return math.Hypot(dx, dy)
}
