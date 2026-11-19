package greyvarserver

import (
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func TestUseState(t *testing.T) {
	for in, want := range map[string]string{"Left": "useLeft", "UpRight": "useUpRight", "": "useDown", "sideways": "useDown"} {
		if got := useState(in); got != want {
			t.Errorf("useState(%q) = %q, want %q", in, got, want)
		}
	}
}

// Pressing the action key shows the use animation to the other players who
// see the player, not to the player (it already played locally).
func TestInteractBroadcastsUse(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	s := newServer()
	me := &Entity{ServerId: 1, Definition: "player", WorldId: "w", GridId: "g"}
	s.entityInstances[me.ServerId] = me
	rp := &RemotePlayer{Username: "me", Entity: me, CurrentWorldId: "w", CurrentGridId: "g", KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	near := &RemotePlayer{Username: "near", CurrentWorldId: "w", CurrentGridId: "g", KnownEntities: map[int64]*Entity{1: me}, currentFrame: &pb.ServerUpdate{}}
	away := &RemotePlayer{Username: "away", CurrentWorldId: "w", CurrentGridId: "other", KnownEntities: map[int64]*Entity{1: me}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"me": rp, "near": near, "away": away}

	processInteractRequest(s, rp, &pb.InteractRequest{Facing: "Left"})

	if got := near.currentFrame.EntityStateChanges; len(got) != 1 || got[0].NewState != "useLeft" || got[0].EntityId != 1 {
		t.Fatalf("near player should see useLeft, got %v", got)
	}
	if len(rp.currentFrame.EntityStateChanges) != 0 || len(away.currentFrame.EntityStateChanges) != 0 {
		t.Fatal("the sender and players on other grids get nothing")
	}
	if me.State != "" {
		t.Fatalf("use is an event; the entity state must not change (got %q)", me.State)
	}
}
