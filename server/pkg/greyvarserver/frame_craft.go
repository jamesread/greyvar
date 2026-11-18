package greyvarserver

import (
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

func processCraftRequest(s *serverInterface, rp *RemotePlayer, req *pb.CraftRequest) {
	if rp == nil || req == nil || req.RecipeId == "" {
		return
	}

	recipe := s.recipeDefinitions[req.RecipeId]
	if recipe == nil {
		msg := "Unknown recipe: " + req.RecipeId
		rp.PendingHudMessage = &msg
		return
	}

	if !rp.canAfford(recipe.Inputs) {
		msg := "Missing ingredients"
		rp.PendingHudMessage = &msg
		return
	}

	if !rp.consumeInputs(recipe.Inputs) {
		msg := "Could not craft"
		rp.PendingHudMessage = &msg
		return
	}

	rp.grantOutputs(recipe.Outputs)
	rp.PendingInventoryUpdate = true

	title := recipe.Title
	if title == "" {
		title = recipe.ID
	}
	text := "Crafted " + title
	rp.PendingHudMessage = &text

	log.WithFields(log.Fields{
		"player": rp.Username,
		"recipe": recipe.ID,
	}).Info("Crafted item")
}
