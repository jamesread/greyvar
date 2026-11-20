package greyvarserver

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesread/greyvar/datlib/itemdefs"
	"github.com/jamesread/greyvar/datlib/rigs"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func newAppearanceTestServer(t *testing.T) (*serverInterface, *RemotePlayer, *RemotePlayer) {
	t.Helper()
	s, a, b := newDisconnectTestServer()
	rig, err := rigs.Load(filepath.Join("..", "..", "..", "res", "img", "textures", "entities", "player", "player.rig.json"))
	if err != nil {
		t.Fatal(err)
	}
	s.rigs = rigs.NewSet(rig)
	s.itemDefinitions = map[string]*itemdefs.ItemDefinition{"key": {}}
	return s, a, b
}

func TestAppearanceRequestBroadcasts(t *testing.T) {
	s, a, b := newAppearanceTestServer(t)
	a.addInventoryItem("key", 1)
	processAppearanceRequest(s, a, &pb.AppearanceRequest{Palettes: map[string]string{"hair": "blonde"}, HeldItem: "key"})

	if a.Entity.Appearance == nil || a.Entity.Appearance.Palettes["hair"] != "blonde" || a.Entity.Appearance.HeldItem != "key" {
		t.Fatalf("appearance not stored: %+v", a.Entity.Appearance)
	}
	if len(b.PendingAppearanceChanges) != 1 || b.PendingAppearanceChanges[0].EntityId != a.Entity.ServerId {
		t.Fatalf("other player not told: %+v", b.PendingAppearanceChanges)
	}
	// Unchanged slots keep the player's username-derived default.
	if got, want := a.Entity.Appearance.Palettes["skin"], s.defaultLook(a.Username).Palettes["skin"]; got != want {
		t.Fatalf("skin changed: %s want %s", got, want)
	}

	b.currentFrame = &pb.ServerUpdate{}
	frameAppearanceChanges(s, b)
	if len(b.currentFrame.EntityAppearanceChanges) != 1 || len(b.PendingAppearanceChanges) != 0 {
		t.Fatal("change not flushed into frame")
	}

	processAppearanceRequest(s, a, &pb.AppearanceRequest{HeldItem: "-"})
	if a.Entity.Appearance.HeldItem != "" || a.Entity.Appearance.Palettes["hair"] != "blonde" {
		t.Fatalf("clearing held item changed other fields: %+v", a.Entity.Appearance)
	}
}

func TestAppearanceRequestRejectsUnknown(t *testing.T) {
	s, a, b := newAppearanceTestServer(t)
	for _, req := range []*pb.AppearanceRequest{
		{Palettes: map[string]string{"hair": "plaid"}},
		{Parts: map[string]string{"hair": "mohawk"}},
	} {
		processAppearanceRequest(s, a, req)
	}
	if a.Entity.Appearance != nil {
		t.Fatalf("invalid requests must not change appearance: %+v", a.Entity.Appearance)
	}
	if len(b.PendingAppearanceChanges) != 0 {
		t.Fatal("invalid requests must not broadcast")
	}
	if len(a.PendingConsoleMessages) != 2 || !strings.Contains(a.PendingConsoleMessages[0], `unknown hair palette "plaid"`) {
		t.Fatalf("want 2 error replies, got %v", a.PendingConsoleMessages)
	}
}

// Held items must be owned; a bad held item drops only itself, and the item
// is put away when it leaves the inventory.
func TestHeldItemFollowsInventory(t *testing.T) {
	s, a, b := newAppearanceTestServer(t)
	processAppearanceRequest(s, a, &pb.AppearanceRequest{Palettes: map[string]string{"hair": "red"}, HeldItem: "key"})
	if a.Entity.Appearance == nil || a.Entity.Appearance.Palettes["hair"] != "red" || a.Entity.Appearance.HeldItem != "" {
		t.Fatalf("unowned key: palette should apply, held item should not: %+v", a.Entity.Appearance)
	}
	if !strings.Contains(strings.Join(a.PendingConsoleMessages, "\n"), "you have no key") {
		t.Fatalf("want a 'you have no key' reply, got %v", a.PendingConsoleMessages)
	}
	processAppearanceRequest(s, a, &pb.AppearanceRequest{HeldItem: "sword"})
	if !strings.Contains(strings.Join(a.PendingConsoleMessages, "\n"), `unknown item "sword"`) {
		t.Fatal("unknown item should be reported")
	}

	a.addInventoryItem("key", 1)
	processAppearanceRequest(s, a, &pb.AppearanceRequest{HeldItem: "key"})
	if a.Entity.Appearance.HeldItem != "key" {
		t.Fatal("owned key should be held")
	}

	b.PendingAppearanceChanges = nil
	a.removeInventoryItem("key", 1)
	a.PendingInventoryUpdate = true
	a.currentFrame = &pb.ServerUpdate{}
	frameInventoryUpdates(s, a)
	if a.Entity.Appearance.HeldItem != "" {
		t.Fatal("held key should be put away once it leaves the inventory")
	}
	if len(b.PendingAppearanceChanges) != 1 || b.PendingAppearanceChanges[0].Appearance.HeldItem != "" {
		t.Fatalf("others should see the key put away: %+v", b.PendingAppearanceChanges)
	}
}

func TestSpawnCarriesAppearance(t *testing.T) {
	s, a, _ := newAppearanceTestServer(t)
	look := s.defaultLook("someone")
	a.Entity.Appearance = &look
	pbLook := appearanceToPB(a.Entity.Appearance)
	if pbLook.Palettes["shirt"] != look.Palettes["shirt"] {
		t.Fatal("appearanceToPB lost palettes")
	}
	if appearanceToPB(nil) != nil {
		t.Fatal("nil appearance must map to nil")
	}
}
