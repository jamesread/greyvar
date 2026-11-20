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

	s.tickCrops(nowMs())

	s.tickNPCs(nowMs())

	s.tickTimers(nowMs())

	s.createServerUpdates();

	s.sendServerUpdates();

	//log.Infof("frame")
}

// maxDeferredMoves caps queued moves per player (see processPlayerRequests).
const maxDeferredMoves = 3

func (s *serverInterface) processPlayerRequests() {
	// Allocate every frame before processing anyone's requests: a request can
	// queue updates for other players (e.g. a bump), which must not be wiped
	// when the loop reaches them.
	for _, p := range s.remotePlayers {
		p.currentFrame = &pb.ServerUpdate{}
	}

	for _, p := range s.remotePlayers {
		// One move per player per frame (that is the walking speed); later
		// moves wait for the next frame rather than being dropped.
		var deferred []*pb.ClientRequests
		moved := false

		for len(p.pendingRequests) > 0 {
			req := p.pendingRequests[0]
			p.pendingRequests = p.pendingRequests[1:]

			if req.MoveRequest != nil {
				if moved {
					deferred = append(deferred, &pb.ClientRequests{MoveRequest: req.MoveRequest})
				} else {
					processMoveRequest(s, p, req.MoveRequest);
					moved = true
				}
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

			if req.DialogChoiceRequest != nil {
				processDialogChoiceRequest(s, p, req.DialogChoiceRequest)
			}

			if req.RiftRequest != nil {
				processRiftRequest(s, p, req.RiftRequest)
			}
		}

		// A client should send about one move per frame; keep a little slack
		// for network bunching but never let a flood build up.
		if len(deferred) > maxDeferredMoves {
			deferred = deferred[len(deferred)-maxDeferredMoves:]
		}
		p.pendingRequests = append(deferred, p.pendingRequests...)

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
		frameTileChanges(s, p)
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

// frameEntityPositions sends the positions that changed since this player
// was last told; everything is sent again after spawning or changing grid.
func frameEntityPositions(s *serverInterface, p *RemotePlayer) {
	if grid := p.CurrentWorldId + "/" + p.CurrentGridId; grid != p.sentPositionsGrid {
		p.sentPositions = nil
		p.sentPositionsGrid = grid
	}
	sent := make(map[int64][2]int32, len(p.sentPositions))
	for _, ent := range s.entitiesOnGrid(p.CurrentWorldId, p.CurrentGridId) {
		pos := [2]int32{ent.X, ent.Y}
		sent[ent.ServerId] = pos
		if last, ok := p.sentPositions[ent.ServerId]; ok && last == pos {
			continue
		}
		p.currentFrame.EntityPositions = append(p.currentFrame.EntityPositions, &pb.EntityPosition{
			X:        ent.X,
			Y:        ent.Y,
			EntityId: ent.ServerId,
		})
	}
	p.sentPositions = sent
}
