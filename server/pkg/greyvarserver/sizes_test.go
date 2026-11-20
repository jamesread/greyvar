package greyvarserver

import (
	"path/filepath"
	"testing"

	"github.com/jamesread/greyvar/datlib/rigs"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// Character sizes: players and NPCs are adults; a look can switch size.

func bothRigs(t *testing.T) *rigs.Set {
	t.Helper()
	ents := filepath.Join("..", "..", "..", "res", "img", "textures", "entities")
	adult, err := rigs.Load(filepath.Join(ents, "adult", "adult.rig.json"))
	if err != nil {
		t.Fatal(err)
	}
	child, err := rigs.Load(filepath.Join(ents, "player", "player.rig.json"))
	if err != nil {
		t.Fatal(err)
	}
	return rigs.NewSet(adult, child)
}

func TestNewPlayersAndNPCsAreAdults(t *testing.T) {
	s, bill := billServer(t)
	if got := s.defaultLook("newcomer").Rig; got != "adult" {
		t.Fatalf("a new player's look is %q, want adult", got)
	}
	want := map[string]map[string]string{
		"bill":      {"hair": "short", "beard": "beard"},
		"geraldine": {"hair": "bun", "outfit": "dress"},
		"ivan":      {"hair": "bald", "beard": "beard"},
		"claire":    {"hair": "bob"},
		"herman":    {"hair": "short"},
	}
	seen := map[string]bool{}
	for _, e := range s.entityInstances {
		parts, ok := want[e.Definition]
		if !ok {
			continue
		}
		seen[e.Definition] = true
		if e.Appearance == nil || e.Appearance.Rig != "adult" {
			t.Errorf("%s is not drawn as an adult: %+v", e.Definition, e.Appearance)
			continue
		}
		for layer, v := range parts {
			if e.Appearance.Parts[layer] != v {
				t.Errorf("%s %s = %q, want %q", e.Definition, layer, e.Appearance.Parts[layer], v)
			}
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("%s not found in the world", name)
		}
	}
	_ = bill
}

func TestAppearanceRequestSwitchesSize(t *testing.T) {
	s, a, _ := newAppearanceTestServer(t)
	s.rigs = bothRigs(t)
	look := s.defaultLook(a.Username)
	look.Parts["hair"], look.Palettes["hair"] = "bun", "red"
	a.Entity.Appearance = &look

	processAppearanceRequest(s, a, &pb.AppearanceRequest{Rig: "child"})
	got := a.Entity.Appearance
	if got.Rig != "child" {
		t.Fatalf("size not switched: %+v", got)
	}
	// The child rig has no bun: that style falls back to the child default,
	// while colours both sizes share are kept.
	if got.Parts["hair"] == "bun" || got.Palettes["hair"] != "red" {
		t.Fatalf("child look should keep the red hair but not the bun: %+v", got)
	}

	processAppearanceRequest(s, a, &pb.AppearanceRequest{Rig: "adult", Parts: map[string]string{"hair": "ponytail", "beard": "beard"}})
	got = a.Entity.Appearance
	if got.Rig != "adult" || got.Parts["hair"] != "ponytail" || got.Parts["beard"] != "beard" || got.Palettes["hair"] != "red" {
		t.Fatalf("back to adult with a ponytail and beard: %+v", got)
	}
}
