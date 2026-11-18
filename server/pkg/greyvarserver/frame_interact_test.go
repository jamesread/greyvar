package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func testOpenGrid(rows, cols uint32) *gridfiles.Grid {
	grid := &gridfiles.Grid{RowCount: rows, ColCount: cols}
	grid.Build()
	return grid
}

func TestInteractHarvestsPotLoot(t *testing.T) {
	grid := testOpenGrid(8, 8)
	s, rp := testServerWithGrid(grid)
	s.remotePlayers = map[string]*RemotePlayer{"p": rp}
	s.entityDefinitions["pot"] = &entdefs.EntityDefinition{
		Title: "pot",
		Loot:  []entdefs.LootDrop{{Item: "shard", Count: 1}},
	}

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	pot := &Entity{ServerId: 2, Definition: "pot", X: 20, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	s.entityInstances[2] = pot
	rp.Entity = player
	rp.KnownEntities = map[int64]*Entity{2: pot}
	rp.currentFrame = &pb.ServerUpdate{}

	processInteractRequest(s, rp, &pb.InteractRequest{EntityId: 2})

	if len(rp.Inventory) != 1 || rp.Inventory[0].Definition != "shard" {
		t.Fatalf("inventory = %#v", rp.Inventory)
	}
	if _, ok := s.entityInstances[2]; ok {
		t.Fatal("pot should be despawned")
	}
}

func TestInteractRejectsOutOfRange(t *testing.T) {
	grid := testOpenGrid(8, 8)
	s, rp := testServerWithGrid(grid)
	s.entityDefinitions["pot"] = &entdefs.EntityDefinition{
		Title: "pot",
		Loot:  []entdefs.LootDrop{{Item: "shard", Count: 1}},
	}

	player := &Entity{ServerId: 1, Definition: "player", X: 0, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	pot := &Entity{ServerId: 2, Definition: "pot", X: 80, Y: 0, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	s.entityInstances[2] = pot
	rp.Entity = player

	processInteractRequest(s, rp, &pb.InteractRequest{EntityId: 2})

	if len(rp.Inventory) != 0 {
		t.Fatalf("expected no harvest out of range, inventory = %#v", rp.Inventory)
	}
}
