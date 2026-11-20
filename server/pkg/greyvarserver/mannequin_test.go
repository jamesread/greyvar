package greyvarserver

import (
	"fmt"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func TestParseWear(t *testing.T) {
	got := parseWear(" outfit=dress, garment = blue ,trim=white,,bad,=x,y= ")
	if len(got) != 3 || got["outfit"] != "dress" || got["garment"] != "blue" || got["trim"] != "white" {
		t.Fatalf("parseWear: %v", got)
	}
	if parseWear("") != nil {
		t.Fatal("empty wear should be nil")
	}
}

// Geraldine's shop has mannequins in real outfits; E puts one on.
func TestMannequinsShowAndGiveTheirOutfit(t *testing.T) {
	s := tailorServer(t)
	const world, grid = "isleOfStarting_dev_tiled", "0.0.tailorShop.tmj"
	var mannequins []*Entity
	for _, e := range s.entitiesOnGrid(world, grid) {
		if e.Definition == "mannequin" {
			mannequins = append(mannequins, e)
		}
	}
	if len(mannequins) != 4 {
		t.Fatalf("want 4 mannequins in the tailor's shop, got %d", len(mannequins))
	}
	var dress, casual, black *Entity
	for _, m := range mannequins {
		a := m.Appearance
		if a == nil || a.Rig != "adult" || a.Parts["form"] != "mannequin" || a.Parts["hair"] != "bald" || m.State != "stand" || a.Palettes["skin"] != "mannequin" {
			t.Fatalf("mannequin look/state: %+v %s", a, m.State)
		}
		if len(entityWear(m)) == 0 {
			t.Fatalf("mannequin at (%d,%d) has no wear", m.X/16, m.Y/16)
		}
		switch {
		case a.Parts["outfit"] == "dress" && a.Palettes["garment"] == "blue" && a.Palettes["trim"] == "white":
			dress = m
		case a.Parts["outfit"] == "none" && a.Palettes["shirt"] == "green" && a.Palettes["pants"] == "khaki":
			casual = m
		case a.Parts["outfit"] == "none" && a.Palettes["shirt"] == "black" && a.Palettes["pants"] == "lightblue":
			black = m
		}
	}
	if dress == nil || casual == nil || black == nil {
		t.Fatal("expected the blue dress and the green shirt/khaki mannequins to wear their outfits")
	}

	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: dress.X, Y: dress.Y + 16, WorldId: world, GridId: grid}
	look := s.defaultLook("guest")
	look.Parts["hair"], look.Palettes["hair"], look.Palettes["skin"] = "bob", "red", "deep"
	player.Appearance = &look
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "guest", Entity: player, CurrentWorldId: world, CurrentGridId: grid,
		KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"guest": rp}

	if !processEntityInteract(s, rp, dress.ServerId) {
		t.Fatal("using the mannequin should do something")
	}
	a := player.Appearance
	if a.Parts["outfit"] != "dress" || a.Palettes["garment"] != "blue" || a.Palettes["trim"] != "white" {
		t.Fatalf("player should now wear the blue dress: %+v", a)
	}
	if a.Parts["form"] == "mannequin" || a.Parts["hair"] != "bob" || a.Palettes["hair"] != "red" || a.Palettes["skin"] != "deep" {
		t.Fatalf("only the clothes change, not the mannequin's form, hair or skin: %+v", a)
	}
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage != "You put on the blue dress." {
		t.Fatalf("message: %v", rp.PendingHudMessage)
	}
	processEntityInteract(s, rp, dress.ServerId)
	if *rp.PendingHudMessage != "You're already wearing the blue dress." {
		t.Fatalf("second use: %q", *rp.PendingHudMessage)
	}

	player.X, player.Y = casual.X, casual.Y+16
	processEntityInteract(s, rp, casual.ServerId)
	a = player.Appearance
	if a.Parts["outfit"] != "none" || a.Palettes["shirt"] != "green" || a.Palettes["pants"] != "khaki" {
		t.Fatalf("the casual mannequin takes the dress off: %+v", a)
	}
}

// The grey mannequin skin is for data only: no new player is given it.
func TestNoPlayerGetsTheMannequinSkin(t *testing.T) {
	s := tailorServer(t)
	for i := 0; i < 500; i++ {
		look := s.defaultLook(fmt.Sprintf("player%d", i))
		if look.Palettes["skin"] == "mannequin" {
			t.Fatalf("player%d got the mannequin skin", i)
		}
	}
}
