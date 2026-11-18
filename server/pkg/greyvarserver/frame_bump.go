package greyvarserver

import pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"

// bumpState names the transient "pushing into a wall" state for a move
// direction, e.g. (-1, 0) -> bumpLeft, (1, -1) -> bumpUpRight. It matches the
// player rig clips (res/sprites/player/clips.yml).
func bumpState(dx, dy int32) string {
	vertical := map[int32]string{-1: "Up", 1: "Down"}[dy]
	horizontal := map[int32]string{-1: "Left", 1: "Right"}[dx]
	if vertical == "" && horizontal == "" {
		return ""
	}
	return "bump" + vertical + horizontal
}

// broadcastBump tells every client that knows the player that it walked into
// something solid. It is an event, not a persistent state: the entity's State
// is unchanged and clients drop back to idle when bumps stop arriving (one is
// sent per blocked move request, ~every 67ms while the key is held).
func (s *serverInterface) broadcastBump(rp *RemotePlayer, dx, dy int32) {
	state := bumpState(dx, dy)
	if state == "" || rp.Entity == nil {
		return
	}
	for _, p := range s.remotePlayers {
		if p.currentFrame == nil {
			continue
		}
		if _, known := p.KnownEntities[rp.Entity.ServerId]; !known && p != rp {
			continue
		}
		p.currentFrame.EntityStateChanges = append(p.currentFrame.EntityStateChanges, &pb.EntityStateChange{
			EntityId: rp.Entity.ServerId,
			NewState: state,
		})
	}
}
