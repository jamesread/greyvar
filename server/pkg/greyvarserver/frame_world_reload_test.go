package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/server/pkg/worlds"
)

func TestReloadWorldRespawnsMapEntities(t *testing.T) {
	grid := &gridfiles.Grid{
		RowCount: 4,
		ColCount: 4,
		Entities: []gridfiles.GridFileEntityInstance{
			{Definition: "pot", Row: 1, Col: 1},
		},
	}
	grid.Build()

	world := &worlds.World{
		ID:        "testWorld",
		SpawnGrid: "testGrid",
		Grids:     map[string]*gridfiles.Grid{"testGrid": grid},
	}

	s := &serverInterface{
		loadedWorlds:      map[string]*worlds.World{"testWorld": world},
		entityInstances:   map[int64]*Entity{},
		entityDefinitions: map[string]*entdefs.EntityDefinition{},
		remotePlayers:     map[string]*RemotePlayer{},
	}
	s.entityDefinitions["pot"] = &entdefs.EntityDefinition{Title: "pot", InitialState: "potlike"}

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	pot := &Entity{ServerId: 2, Definition: "pot", X: 16, Y: 16, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	s.entityInstances[2] = pot

	rp := &RemotePlayer{
		Username:       "p",
		CurrentWorldId: "testWorld",
		CurrentGridId:  "testGrid",
		Entity:         player,
		KnownEntities:  map[int64]*Entity{2: pot},
		currentFrame:   &pb.ServerUpdate{},
	}
	s.remotePlayers["p"] = rp

	s.purgeWorldEntities("testWorld")
	s.mountWorld(world)

	if _, ok := s.entityInstances[2]; ok {
		t.Fatal("old pot entity should be removed")
	}

	potCount := 0
	for _, ent := range s.entityInstances {
		if ent.Definition == "pot" {
			potCount++
		}
	}
	if potCount != 1 {
		t.Fatalf("expected one pot after reload, got %d", potCount)
	}
}
