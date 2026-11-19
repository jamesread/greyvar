package greyvarserver

import pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"

var useFacings = map[string]bool{
	"Up": true, "Down": true, "Left": true, "Right": true,
	"UpLeft": true, "UpRight": true, "DownLeft": true, "DownRight": true,
}

// useState names the one-off "use" event for a facing, e.g. Left -> useLeft
// (the player rig's use clips; diagonals fall back to the side clip on the
// client). Unknown facings use the front view.
func useState(facing string) string {
	if !useFacings[facing] {
		facing = "Down"
	}
	return "use" + facing
}

// broadcastUse tells the other players who can see rp that it pressed the
// action key, so they play its use animation. Like bumps it is an event, not
// a state; the sender already played it locally when the key went down.
func (s *serverInterface) broadcastUse(rp *RemotePlayer, facing string) {
	if rp.Entity == nil {
		return
	}
	change := &pb.EntityStateChange{EntityId: rp.Entity.ServerId, NewState: useState(facing)}
	for _, p := range s.remotePlayers {
		if p == rp || p.currentFrame == nil {
			continue
		}
		if p.CurrentWorldId != rp.CurrentWorldId || p.CurrentGridId != rp.CurrentGridId {
			continue
		}
		if _, known := p.KnownEntities[rp.Entity.ServerId]; !known {
			continue
		}
		p.currentFrame.EntityStateChanges = append(p.currentFrame.EntityStateChanges, change)
	}
}
