package greyvarserver

import (
	"github.com/jamesread/greyvar/datlib/itemdefs"
	"github.com/jamesread/greyvar/datlib/recipes"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

func (s *serverInterface) loadItemCatalog() {
	items, err := itemdefs.LoadAllItemdefs()
	if err != nil {
		log.Warnf("Cannot load itemdefs: %v", err)
		items = map[string]*itemdefs.ItemDefinition{}
	}
	s.itemDefinitions = items
	log.Infof("Loaded %d item definitions", len(items))

	recipeMap, err := recipes.LoadAllRecipes()
	if err != nil {
		log.Warnf("Cannot load recipes: %v", err)
		recipeMap = map[string]*recipes.RecipeDefinition{}
	}
	s.recipeDefinitions = recipeMap
	log.Infof("Loaded %d recipes", len(recipeMap))
}

func frameItemDefinitions(s *serverInterface, p *RemotePlayer) {
	if p.KnownItemdefs == nil {
		p.KnownItemdefs = make(map[string]bool)
	}

	for id, item := range s.itemDefinitions {
		if p.KnownItemdefs[id] {
			continue
		}
		def := &pb.ItemDefinition{
			Id:       item.ID,
			Title:    item.Title,
			StackMax: item.StackMax,
			Category: item.Category,
			Icon:     item.Icon,
		}
		if item.Held != nil {
			def.HeldSprite = item.Held.Sprite
			def.HeldGripX = item.Held.Grip[0]
			def.HeldGripY = item.Held.Grip[1]
		}
		p.currentFrame.ItemDefinitions = append(p.currentFrame.ItemDefinitions, def)
		p.KnownItemdefs[id] = true
	}
}

func frameRecipeDefinitions(s *serverInterface, p *RemotePlayer) {
	if p.KnownRecipes == nil {
		p.KnownRecipes = make(map[string]bool)
	}

	for id, recipe := range s.recipeDefinitions {
		if p.KnownRecipes[id] {
			continue
		}
		net := &pb.RecipeDefinition{
			Id:    recipe.ID,
			Title: recipe.Title,
		}
		for _, input := range recipe.Inputs {
			net.Inputs = append(net.Inputs, &pb.RecipeInput{
				Item:  input.Item,
				Count: input.Count,
			})
		}
		for _, output := range recipe.Outputs {
			net.Outputs = append(net.Outputs, &pb.RecipeOutput{
				Item:  output.Item,
				Count: output.Count,
			})
		}
		p.currentFrame.RecipeDefinitions = append(p.currentFrame.RecipeDefinitions, net)
		p.KnownRecipes[id] = true
	}
}

func (s *serverInterface) resolvePickupItemID(ent *Entity) string {
	if ent == nil {
		return ""
	}

	entdef := s.entityDefinitions[ent.Definition]
	if entdef != nil && entdef.ItemID != "" {
		return entdef.ItemID
	}

	vis, _, ok := s.entityVisualInfo(ent)
	if ok {
		if id := entityProp(ent, vis, "item_id"); id != "" {
			return id
		}
	}

	return ent.Definition
}

func itemTitle(s *serverInterface, itemID string) string {
	if item := s.itemDefinitions[itemID]; item != nil && item.Title != "" {
		return item.Title
	}
	return itemID
}
