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

	ent := s.entityInstances[req.EntityId]
	if ent == nil {
		return
	}

	if ent.WorldId != rp.CurrentWorldId || ent.GridId != rp.CurrentGridId {
		return
	}

	if ent.ServerId == rp.Entity.ServerId {
		return
	}

	if !entityInInteractRange(rp.Entity, ent) {
		return
	}

	entdef := s.entityDefinitions[ent.Definition]
	if entdef == nil || len(entdef.Loot) == 0 {
		return
	}

	log.WithFields(log.Fields{
		"player":   rp.Username,
		"entityId": ent.ServerId,
		"type":     ent.Definition,
	}).Info("Interact harvest")

	s.harvestEntity(rp, ent)
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
	if entdef == nil || len(entdef.Loot) == 0 {
		return
	}

	for _, drop := range entdef.Loot {
		count := drop.Count
		if count <= 0 {
			count = 1
		}
		rp.addInventoryItem(drop.Item, count)
	}
	rp.PendingInventoryUpdate = true

	label := ent.Definition
	if len(entdef.Loot) == 1 {
		label = itemTitle(s, entdef.Loot[0].Item)
	}
	text := "Collected " + label
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
