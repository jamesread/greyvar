package greyvarserver

import (
	"github.com/jamesread/greyvar/datlib/recipes"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

const defaultStackMax = 99

func (rp *RemotePlayer) inventoryCount(definition string) int32 {
	if rp == nil || definition == "" {
		return 0
	}
	for _, item := range rp.Inventory {
		if item.Definition == definition {
			return item.Count
		}
	}
	return 0
}

func (rp *RemotePlayer) removeInventoryItem(definition string, count int32) bool {
	if rp == nil || definition == "" || count <= 0 {
		return false
	}

	for i := range rp.Inventory {
		if rp.Inventory[i].Definition != definition {
			continue
		}
		if rp.Inventory[i].Count < count {
			return false
		}
		rp.Inventory[i].Count -= count
		if rp.Inventory[i].Count == 0 {
			rp.Inventory = append(rp.Inventory[:i], rp.Inventory[i+1:]...)
		}
		return true
	}

	return false
}

func (rp *RemotePlayer) canAfford(inputs []recipes.RecipeInput) bool {
	for _, input := range inputs {
		need := input.Count
		if need <= 0 {
			need = 1
		}
		if rp.inventoryCount(input.Item) < need {
			return false
		}
	}
	return true
}

func (rp *RemotePlayer) consumeInputs(inputs []recipes.RecipeInput) bool {
	if !rp.canAfford(inputs) {
		return false
	}
	for _, input := range inputs {
		need := input.Count
		if need <= 0 {
			need = 1
		}
		if !rp.removeInventoryItem(input.Item, need) {
			return false
		}
	}
	return true
}

func (rp *RemotePlayer) grantOutputs(outputs []recipes.RecipeOutput) {
	for _, output := range outputs {
		count := output.Count
		if count <= 0 {
			count = 1
		}
		rp.addInventoryItem(output.Item, count)
	}
}

func inventoryToProto(items []InventoryItem) *pb.InventoryUpdate {
	update := &pb.InventoryUpdate{}
	for _, item := range items {
		update.Items = append(update.Items, &pb.InventoryItem{
			Definition: item.Definition,
			Count:      item.Count,
		})
	}
	return update
}
