package greyvarserver

import (
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func TestBumpState(t *testing.T) {
	cases := map[[2]int32]string{
		{-1, 0}: "bumpLeft", {1, 0}: "bumpRight", {0, -1}: "bumpUp", {0, 1}: "bumpDown",
		{-1, -1}: "bumpUpLeft", {1, -1}: "bumpUpRight", {-1, 1}: "bumpDownLeft", {1, 1}: "bumpDownRight",
		{0, 0}: "",
	}
	for d, want := range cases {
		if got := bumpState(d[0], d[1]); got != want {
			t.Errorf("bumpState%v = %q, want %q", d, got, want)
		}
	}
}

func TestBroadcastBumpReachesKnowingPlayers(t *testing.T) {
	s, a, b := newDisconnectTestServer()
	a.currentFrame = &pb.ServerUpdate{}
	b.currentFrame = &pb.ServerUpdate{}
	stranger := &RemotePlayer{Username: "c", KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers["c"] = stranger

	s.broadcastBump(a, 1, 0)

	for _, p := range []*RemotePlayer{a, b} {
		ch := p.currentFrame.EntityStateChanges
		if len(ch) != 1 || ch[0].EntityId != a.Entity.ServerId || ch[0].NewState != "bumpRight" {
			t.Fatalf("%s: want one bumpRight for %d, got %+v", p.Username, a.Entity.ServerId, ch)
		}
	}
	if len(stranger.currentFrame.EntityStateChanges) != 0 {
		t.Fatal("player who does not know the entity must not get the bump")
	}
	if a.Entity.State == "bumpRight" {
		t.Fatal("bump must not change the persistent entity state")
	}
}
