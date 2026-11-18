package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/recipes"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func TestCraftConsumesInputsAndGrantsOutput(t *testing.T) {
	s := &serverInterface{
		recipeDefinitions: map[string]*recipes.RecipeDefinition{
			"craft_gem": {
				ID: "craft_gem",
				Inputs: []recipes.RecipeInput{
					{Item: "shard", Count: 2},
				},
				Outputs: []recipes.RecipeOutput{
					{Item: "gem", Count: 1},
				},
			},
		},
	}
	rp := &RemotePlayer{
		Username: "p",
		Inventory: []InventoryItem{
			{Definition: "shard", Count: 2},
		},
		currentFrame: &pb.ServerUpdate{},
	}

	processCraftRequest(s, rp, &pb.CraftRequest{RecipeId: "craft_gem"})

	if rp.inventoryCount("shard") != 0 {
		t.Fatalf("shards not consumed: %#v", rp.Inventory)
	}
	if rp.inventoryCount("gem") != 1 {
		t.Fatalf("gem not granted: %#v", rp.Inventory)
	}
	if !rp.PendingInventoryUpdate {
		t.Fatal("expected inventory update pending")
	}
}
