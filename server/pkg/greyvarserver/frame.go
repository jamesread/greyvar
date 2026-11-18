package greyvarserver;

import (
	"context"
	"time"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
	"github.com/coder/websocket/wsjson"
)

func (s *serverInterface) frame() {
	s.frameTime = time.Now().UnixNano();

	s.processPlayerRequests();

	s.createServerUpdates();

	s.sendServerUpdates();

	//log.Infof("frame")
}

func (s *serverInterface) processPlayerRequests() {
	// Allocate every frame before processing anyone's requests: a request can
	// queue updates for other players (e.g. a bump), which must not be wiped
	// when the loop reaches them.
	for _, p := range s.remotePlayers {
		p.currentFrame = &pb.ServerUpdate{}
	}

	for _, p := range s.remotePlayers {

		for len(p.pendingRequests) > 0 {
			req := p.pendingRequests[0]
			p.pendingRequests = p.pendingRequests[1:]

			if req.MoveRequest != nil {
				processMoveRequest(s, p, req.MoveRequest);
			}

			if req.WorldListRequest != nil {
				processWorldListRequest(s, p)
			}

			if req.WorldLoadRequest != nil {
				processWorldLoadRequest(s, p, req.WorldLoadRequest)
			}

			if req.InteractRequest != nil {
				processInteractRequest(s, p, req.InteractRequest)
			}

			if req.CraftRequest != nil {
				processCraftRequest(s, p, req.CraftRequest)
			}

			if req.WorldReloadRequest != nil {
				processWorldReloadRequest(s, p, req.WorldReloadRequest)
			}

			if req.AppearanceRequest != nil {
				processAppearanceRequest(s, p, req.AppearanceRequest)
			}
		}

	}
}

func (s *serverInterface) createServerUpdates() {
	for _, p := range s.remotePlayers {
		frameNewEntdefs(s, p)
		frameSpawnPlayer(s, p)
		frameEntityDespawns(s, p)
		frameGridUpdates(s, p)
		frameSpawnEntities(s, p)
		frameEntityPositions(s, p);
		frameAppearanceChanges(s, p)
		frameConsoleMessages(s, p)
		frameHudMessages(s, p)
		frameInventoryUpdates(s, p)
		frameItemDefinitions(s, p)
		frameRecipeDefinitions(s, p)
	}
}

func (s *serverInterface) sendServerUpdates() {
	for _, player := range s.remotePlayers {
		s.sendServerFrameForPlayer(player)
	}
}

func (s *serverInterface) sendServerFrameForPlayer(p *RemotePlayer) {
	s.sendServerFrame(p.currentFrame, p)
}

func (s *serverInterface) sendServerFrame(frame *pb.ServerUpdate, p *RemotePlayer) {
	err := wsjson.Write(context.Background(), p.Connection, frame)

	if err != nil {
		log.Errorf("Could not marshal obj to protobuf in sendMessage: %v", err);
		return
	}
}

func frameEntityDespawns(s *serverInterface, p *RemotePlayer) {
	if len(p.PendingDespawns) == 0 {
		return
	}

	p.currentFrame.EntityDespawns = append(p.currentFrame.EntityDespawns, p.PendingDespawns...)
	p.PendingDespawns = nil
}

func frameEntityPositions(s *serverInterface, p *RemotePlayer) {
	for _, ent := range s.entitiesOnGrid(p.CurrentWorldId, p.CurrentGridId) {
		entpos := pb.EntityPosition {
			X: ent.X,
			Y: ent.Y,
			EntityId: ent.ServerId,
		}

		p.currentFrame.EntityPositions = append(p.currentFrame.EntityPositions, &entpos);
	}
}
