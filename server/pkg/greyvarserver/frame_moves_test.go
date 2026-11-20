package greyvarserver

import (
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// Moves are applied one per frame and never dropped: when two arrive in the
// same frame the second waits for the next (it used to be thrown away by a
// throttle, so a third of all steps were lost and walking juddered).
func TestMovesOnePerFrameNoneDropped(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	saved := playerHitbox // newServer sets it from the rig; other tests expect the default
	defer func() { playerHitbox = saved }()
	s := newServer()
	const world, grid = "isleOfStarting_dev_tiled", "1.2.tmj"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 112, Y: 128, WorldId: world, GridId: grid}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "walker", Entity: player, CurrentWorldId: world, CurrentGridId: grid,
		KnownEntities: map[int64]*Entity{}, KnownEntdefs: map[string]bool{}}
	s.remotePlayers = map[string]*RemotePlayer{"walker": rp}
	right := func() *pb.ClientRequests { return &pb.ClientRequests{MoveRequest: &pb.MoveRequest{X: 1}} }

	// The client sends one move per frame (15Hz); network jitter bunches
	// them: two in one frame, none in the next. 18 moves in all.
	var xs []int32
	arrivals := []int{1, 2, 0, 1, 1, 2, 0, 2, 0, 1, 1, 1, 2, 0, 1, 2, 0, 1}
	for _, n := range arrivals {
		for k := 0; k < n; k++ {
			rp.pendingRequests = append(rp.pendingRequests, right())
		}
		s.processPlayerRequests()
		xs = append(xs, player.X)
	}
	for len(rp.pendingRequests) > 0 {
		s.processPlayerRequests()
		xs = append(xs, player.X)
	}
	if player.X != 112+18*DISTANCE_PER_REQUEST {
		t.Fatalf("all 18 moves should land (x=%d), positions %v", player.X, xs)
	}
	for i := 1; i < len(xs); i++ {
		if d := xs[i] - xs[i-1]; d != DISTANCE_PER_REQUEST {
			t.Fatalf("one 6px step per frame, got %d at frame %d: %v", d, i, xs)
		}
	}

	// A flood is capped, not queued forever.
	for i := 0; i < 50; i++ {
		rp.pendingRequests = append(rp.pendingRequests, right())
	}
	s.processPlayerRequests()
	if len(rp.pendingRequests) > maxDeferredMoves {
		t.Fatalf("queued moves should be capped at %d, have %d", maxDeferredMoves, len(rp.pendingRequests))
	}
}
