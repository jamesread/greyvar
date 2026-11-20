package greyvarserver

import (
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// Positions go out when they change (and again after a spawn or grid change),
// not every frame for every entity.
func TestEntityPositionsOnlyWhenChanged(t *testing.T) {
	s := &serverInterface{entityInstances: map[int64]*Entity{}}
	a := &Entity{ServerId: 1, X: 10, Y: 10, WorldId: "w", GridId: "g"}
	b := &Entity{ServerId: 2, X: 50, Y: 50, WorldId: "w", GridId: "g"}
	s.entityInstances[1], s.entityInstances[2] = a, b
	p := &RemotePlayer{CurrentWorldId: "w", CurrentGridId: "g"}
	frame := func() []int64 {
		p.currentFrame = &pb.ServerUpdate{}
		frameEntityPositions(s, p)
		var ids []int64
		for _, ep := range p.currentFrame.EntityPositions {
			ids = append(ids, ep.EntityId)
		}
		return ids
	}
	if got := frame(); len(got) != 2 {
		t.Fatalf("first frame sends everyone, got %v", got)
	}
	if got := frame(); len(got) != 0 {
		t.Fatalf("nothing moved, nothing sent, got %v", got)
	}
	a.X += 6
	if got := frame(); len(got) != 1 || got[0] != 1 {
		t.Fatalf("only the mover, got %v", got)
	}
	delete(p.sentPositions, 2) // as when (re)spawned for this player
	if got := frame(); len(got) != 1 || got[0] != 2 {
		t.Fatalf("a respawned entity's position is sent again, got %v", got)
	}
	a.GridId, b.GridId = "g2", "g2"
	p.CurrentGridId = "g2"
	if got := frame(); len(got) != 2 {
		t.Fatalf("after a grid change everything is sent again, got %v", got)
	}
}
