package greyvarserver

import (
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

const interactRangePx = 24

func processInteractRequest(s *serverInterface, rp *RemotePlayer, req *pb.InteractRequest) {
	if rp == nil || req == nil || rp.Entity == nil {
		return
	}
	s.broadcastUse(rp, req.GetFacing())
	if processEntityInteract(s, rp, req.EntityId) {
		return
	}
	// Nothing to do with the entity (or none sent): act on the faced cell.
	if cell := req.GetCell(); cell != nil {
		s.processCellInteract(rp, cell.GetRow(), cell.GetCol())
	}
}

// processEntityInteract harvests a lootable entity in range. Returns whether
// it did anything.
func processEntityInteract(s *serverInterface, rp *RemotePlayer, entityId int64) bool {
	ent := s.entityInstances[entityId]
	if ent == nil {
		return false
	}

	if ent.WorldId != rp.CurrentWorldId || ent.GridId != rp.CurrentGridId {
		return false
	}

	if ent.ServerId == rp.Entity.ServerId {
		return false
	}

	if !entityInInteractRange(rp.Entity, ent) {
		return false
	}

	entdef := s.entityDefinitions[ent.Definition]
	if entdef == nil {
		return false
	}
	// Entities with interact rules (levers, chests) never fall through to
	// loot harvesting: an opened chest must not hand out its loot again.
	if entdef.Interact != nil {
		return s.interactBehaviour(rp, ent, entdef)
	}
	if len(lootFor(ent, entdef)) == 0 {
		return false
	}

	log.WithFields(log.Fields{
		"player":   rp.Username,
		"entityId": ent.ServerId,
		"type":     ent.Definition,
	}).Info("Interact harvest")

	s.harvestEntity(rp, ent)
	return true
}

func entityInInteractRange(player, target *Entity) bool {
	if player == nil || target == nil {
		return false
	}
	dx := absInt32(player.X - target.X)
	dy := absInt32(player.Y - target.Y)
	return dx <= interactRangePx && dy <= interactRangePx
}

func (s *serverInterface) harvestEntity(rp *RemotePlayer, ent *Entity) {
	entdef := s.entityDefinitions[ent.Definition]
	loot := lootFor(ent, entdef)
	if len(loot) == 0 {
		return
	}

	text := "Collected " + s.grantLoot(rp, loot)
	rp.PendingHudMessage = &text
	rp.PendingHudEntityId = ent.ServerId

	s.despawnEntity(ent.ServerId)
}

func (s *serverInterface) despawnEntity(entityID int64) {
	delete(s.entityInstances, entityID)
	for _, other := range s.remotePlayers {
		if _, known := other.KnownEntities[entityID]; known {
			other.PendingDespawns = append(other.PendingDespawns, entityID)
			delete(other.KnownEntities, entityID)
		}
		if other.ActiveEntityIds != nil {
			delete(other.ActiveEntityIds, entityID)
		}
	}
}
