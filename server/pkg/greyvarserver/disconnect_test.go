package greyvarserver

import "testing"

func newDisconnectTestServer() (*serverInterface, *RemotePlayer, *RemotePlayer) {
	s := &serverInterface{
		entityInstances: map[int64]*Entity{},
		remotePlayers:   map[string]*RemotePlayer{},
	}

	join := func(id int64, grid string) *RemotePlayer {
		ent := &Entity{ServerId: id, Definition: "player", WorldId: "w", GridId: grid}
		s.entityInstances[id] = ent
		rp := &RemotePlayer{
			Username:       "bob", // all clients currently share a display name
			Entity:         ent,
			CurrentWorldId: "w",
			CurrentGridId:  grid,
			KnownEntities:  map[int64]*Entity{},
		}
		s.remotePlayers[playerKey(rp)] = rp
		return rp
	}

	a := join(1, "0.0.tmj")
	b := join(2, "0.0.tmj")
	a.KnownEntities[2] = b.Entity
	b.KnownEntities[1] = a.Entity
	return s, a, b
}

// Two clients with the same username must both stay registered, and each
// disconnect must remove the right player and despawn it for the other.
func TestRemovePlayerWithSharedUsername(t *testing.T) {
	s, a, b := newDisconnectTestServer()

	if len(s.remotePlayers) != 2 {
		t.Fatalf("both players should be registered, got %d", len(s.remotePlayers))
	}

	s.removePlayer(a)

	if _, ok := s.entityInstances[1]; ok {
		t.Fatal("disconnected player's entity is still in the world")
	}
	if _, ok := s.entityInstances[2]; !ok {
		t.Fatal("the other player's entity was removed")
	}
	if len(s.remotePlayers) != 1 || s.remotePlayers[playerKey(b)] != b {
		t.Fatalf("expected only player b to remain, got %v", s.remotePlayers)
	}
	if len(b.PendingDespawns) != 1 || b.PendingDespawns[0] != 1 {
		t.Fatalf("player b should be told to despawn entity 1, got %v", b.PendingDespawns)
	}
	if _, known := b.KnownEntities[1]; known {
		t.Fatal("player b still tracks the removed entity")
	}
}

// A player who saw the entity, even on another grid, must still be told.
func TestRemovePlayerDespawnsForAnyoneWhoKnowsIt(t *testing.T) {
	s, a, b := newDisconnectTestServer()
	b.CurrentGridId = "0.1.tmj"

	s.removePlayer(a)

	if len(b.PendingDespawns) != 1 {
		t.Fatalf("expected despawn for a viewer on another grid, got %v", b.PendingDespawns)
	}
}

func TestRemovePlayerTwiceIsHarmless(t *testing.T) {
	s, a, b := newDisconnectTestServer()
	s.removePlayer(a)
	s.removePlayer(a)

	if len(b.PendingDespawns) != 1 {
		t.Fatalf("second remove should not queue another despawn, got %v", b.PendingDespawns)
	}
}
