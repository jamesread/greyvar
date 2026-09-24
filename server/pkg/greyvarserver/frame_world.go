package greyvarserver

import (
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/server/pkg/worlds"
	log "github.com/sirupsen/logrus"
)

func processWorldListRequest(s *serverInterface, rp *RemotePlayer) {
	summaries, err := worlds.ListWorlds()
	if err != nil {
		rp.PendingConsoleMessages = append(rp.PendingConsoleMessages, "world-list failed: "+err.Error())
		return
	}

	response := &pb.WorldListResponse{}
	for _, summary := range summaries {
		response.Worlds = append(response.Worlds, &pb.WorldSummary{
			WorldId:   summary.ID,
			Title:     summary.Title,
			GridCount: uint32(summary.GridCount),
			SpawnGrid: summary.SpawnGrid,
		})
	}

	rp.currentFrame.WorldListResponse = response
}

func processWorldLoadRequest(s *serverInterface, rp *RemotePlayer, req *pb.WorldLoadRequest) {
	worldId := req.GetWorldId()
	if worldId == "" {
		rp.PendingConsoleMessages = append(rp.PendingConsoleMessages, "world-load failed: world id required")
		return
	}

	if err := s.loadPlayerWorld(rp, worldId); err != nil {
		rp.PendingConsoleMessages = append(rp.PendingConsoleMessages, "world-load failed: "+err.Error())
		log.WithFields(log.Fields{
			"player": rp.Username,
			"world":  worldId,
			"error":  err.Error(),
		}).Warn("world-load failed")
		return
	}

	rp.PendingConsoleMessages = append(rp.PendingConsoleMessages,
		"loaded world "+worldId+" (grid "+rp.CurrentGridId+")")
}

func (s *serverInterface) loadPlayerWorld(rp *RemotePlayer, worldId string) error {
	world, err := s.ensureWorldLoaded(worldId)
	if err != nil {
		return err
	}

	s.leaveCurrentLocation(rp)

	spawnGridId, spawnX, spawnY, ok := worlds.PlayerSpawnLocation(world)
	if !ok {
		return errSpawnGridMissing
	}

	if world.Grids[spawnGridId] == nil {
		return errSpawnGridMissing
	}
	rp.CurrentWorldId = worldId
	rp.CurrentGridId = spawnGridId
	rp.Entity.WorldId = worldId
	rp.Entity.GridId = spawnGridId
	rp.Entity.X = spawnX
	rp.Entity.Y = spawnY
	rp.NeedsGridUpdate = true

	log.WithFields(log.Fields{
		"player": rp.Username,
		"world":  worldId,
		"grid":   spawnGridId,
	}).Info("Player loaded world")

	return nil
}

func (s *serverInterface) leaveCurrentLocation(rp *RemotePlayer) {
	oldGridId := rp.CurrentGridId
	oldWorldId := rp.CurrentWorldId

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
		if rp.Entity != nil && id == rp.Entity.ServerId {
			continue
		}
		rp.PendingDespawns = append(rp.PendingDespawns, id)
	}

	clearSignHud(rp)
	rp.ActiveEntityIds = make(map[int64]bool)
}

func frameConsoleMessages(s *serverInterface, p *RemotePlayer) {
	if len(p.PendingConsoleMessages) == 0 {
		return
	}

	p.currentFrame.ConsoleMessage = &pb.ConsoleMessage{
		Text: p.PendingConsoleMessages[0],
	}
	p.PendingConsoleMessages = p.PendingConsoleMessages[1:]
}

func frameHudMessages(s *serverInterface, p *RemotePlayer) {
	if p.PendingHudMessage == nil {
		return
	}
	p.currentFrame.HudMessage = &pb.HudMessage{
		Text:     *p.PendingHudMessage,
		EntityId: p.PendingHudEntityId,
	}
	p.PendingHudMessage = nil
	p.PendingHudEntityId = 0
}

func frameInventoryUpdates(s *serverInterface, p *RemotePlayer) {
	if !p.PendingInventoryUpdate {
		return
	}
	items := make([]*pb.InventoryItem, 0, len(p.Inventory))
	for _, it := range p.Inventory {
		items = append(items, &pb.InventoryItem{
			Definition: it.Definition,
			Count:      it.Count,
		})
	}
	p.currentFrame.InventoryUpdate = &pb.InventoryUpdate{Items: items}
	p.PendingInventoryUpdate = false
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
