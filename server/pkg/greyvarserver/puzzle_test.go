package greyvarserver

import (
	"strings"
	"testing"

	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/itemdefs"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func newPuzzleServer(t *testing.T) (*serverInterface, *RemotePlayer, *RemotePlayer) {
	t.Helper()
	blockingTileTextures = map[string]bool{}
	grid := &gridfiles.Grid{RowCount: 4, ColCount: 6}
	grid.Build()
	for r := uint32(0); r < 4; r++ {
		for c := uint32(0); c < 6; c++ {
			grid.Tiles[r][c] = &gridfiles.Tile{Row: r, Col: c, Texture: "floor.png", Traversable: true, CollisionFromTSX: true}
		}
	}
	s, rp := testServerWithGrid(grid)
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.itemDefinitions = map[string]*itemdefs.ItemDefinition{
		"rusty_key": {ID: "rusty_key", Title: "Rusty Key"}, "note": {ID: "note", Title: "Note"}, "shard": {ID: "shard", Title: "Shard"},
	}
	s.entityDefinitions["wallSwitch"] = &entdefs.EntityDefinition{Title: "wallSwitch", Solid: true, InitialState: "flickedLeft",
		Interact: &entdefs.Interact{Toggle: []string{"flickedLeft", "flickedRight"}}}
	s.entityDefinitions["cellDoor"] = &entdefs.EntityDefinition{Title: "cellDoor", Solid: true, InitialState: "closed", PassableStates: []string{"open"}}
	s.entityDefinitions["chest"] = &entdefs.EntityDefinition{Title: "chest", Solid: true, InitialState: "closed", Interact: &entdefs.Interact{Open: "open", Close: "shut"}}
	s.entityDefinitions["pot"] = &entdefs.EntityDefinition{Title: "pot", Solid: true, InitialState: "potlike", Loot: []entdefs.LootDrop{{Item: "shard", Count: 1}}}

	add := func(id int64, def string, col int32, props map[string]string) {
		s.entityInstances[id] = &Entity{ServerId: id, Definition: def, State: s.entityDefinitions[def].InitialState,
			X: col * 16, Y: 0, WorldId: "testWorld", GridId: "testGrid", Properties: props}
	}
	add(10, "wallSwitch", 0, map[string]string{"id": "sw1"})
	add(11, "wallSwitch", 1, map[string]string{"id": "sw2"})
	add(12, "cellDoor", 2, map[string]string{"open_when": "sw1=flickedRight, sw2=flickedLeft"})
	add(13, "chest", 3, map[string]string{"requires": "rusty_key", "loot": "note", "msg": "A note: right, left.", "locked_msg": "Locked tight."})
	add(14, "pot", 4, map[string]string{"loot": "rusty_key"})

	player := &Entity{ServerId: 1, Definition: "player", X: 16, Y: 16, WorldId: "testWorld", GridId: "testGrid"}
	s.entityInstances[1] = player
	rp.Entity, rp.Username = player, "p1"
	rp.currentFrame = &pb.ServerUpdate{}
	rp.KnownEntities = map[int64]*Entity{}
	other := &RemotePlayer{Username: "p2", CurrentWorldId: "testWorld", CurrentGridId: "testGrid", currentFrame: &pb.ServerUpdate{}, KnownEntities: map[int64]*Entity{}}
	for id, e := range s.entityInstances {
		rp.KnownEntities[id], other.KnownEntities[id] = e, e
	}
	s.remotePlayers = map[string]*RemotePlayer{"p1": rp, "p2": other}
	s.evaluateOpenWhen("testWorld", "testGrid", 0)
	return s, rp, other
}

func TestParseOpenWhenAndLoot(t *testing.T) {
	conds, err := parseOpenWhen(" a=on , b = off ")
	if err != nil || len(conds) != 2 || conds[1] != (condition{"b", "off"}) {
		t.Fatalf("parse: %v %v", conds, err)
	}
	for _, bad := range []string{"", "a", "a=", "=on"} {
		if _, err := parseOpenWhen(bad); err == nil {
			t.Errorf("%q should not parse", bad)
		}
	}
	loot := parseLoot("rusty_key, shard:3, ")
	if len(loot) != 2 || loot[0].Item != "rusty_key" || loot[0].Count != 1 || loot[1].Count != 3 {
		t.Fatalf("loot = %+v", loot)
	}
}

func TestLeversOpenTheDoor(t *testing.T) {
	s, rp, other := newPuzzleServer(t)
	door := s.entityInstances[12]
	if door.State != "closed" {
		t.Fatal("door should start closed")
	}
	if blocked, _ := s.isBlockedBySolidEntity(rp, rp.Entity, 32, 0); !blocked {
		t.Fatal("closed door must block")
	}

	// sw1 -> flickedRight (sw2 already flickedLeft) opens the door.
	if !processEntityInteract(s, rp, 10) || s.entityInstances[10].State != "flickedRight" {
		t.Fatal("interact should flip sw1")
	}
	if door.State != "open" {
		t.Fatal("door should open when all conditions hold")
	}
	if blocked, _ := s.isBlockedBySolidEntity(rp, rp.Entity, 32, 0); blocked {
		t.Fatal("open door must not block")
	}
	var sawDoor bool
	for _, ch := range other.currentFrame.EntityStateChanges {
		sawDoor = sawDoor || (ch.EntityId == 12 && ch.NewState == "open")
	}
	if !sawDoor {
		t.Fatal("other players on the grid should see the door open")
	}

	// Flipping sw2 breaks the combination: door closes again.
	processEntityInteract(s, rp, 11)
	if door.State != "closed" {
		t.Fatal("door should close when a condition fails")
	}
}

func TestLockedChestAndPotKey(t *testing.T) {
	s, rp, _ := newPuzzleServer(t)
	chest := s.entityInstances[13]
	rp.Entity.X = 48 // next to the chest

	processEntityInteract(s, rp, 13)
	if chest.State != "closed" || *rp.PendingHudMessage != "Locked tight." {
		t.Fatalf("chest should stay locked without the key: %s %v", chest.State, *rp.PendingHudMessage)
	}

	rp.Entity.X = 64 // next to the pot hiding the key
	if !processEntityInteract(s, rp, 14) || rp.inventoryCount("rusty_key") != 1 || rp.inventoryCount("shard") != 0 {
		t.Fatal("the pot's loot property should give the key instead of a shard")
	}
	if s.entityInstances[14] != nil {
		t.Fatal("harvested pot should be gone")
	}

	rp.Entity.X = 48
	processEntityInteract(s, rp, 13)
	if chest.State != "open" || rp.inventoryCount("rusty_key") != 0 || rp.inventoryCount("note") != 1 {
		t.Fatalf("chest should open, use up the key and give the note: %s key=%d note=%d", chest.State, rp.inventoryCount("rusty_key"), rp.inventoryCount("note"))
	}
	if !strings.Contains(*rp.PendingHudMessage, "right, left") {
		t.Fatalf("opening should show the chest's message, got %q", *rp.PendingHudMessage)
	}

	// Close it again, then reopen: no key needed, no second note, no message.
	rp.PendingHudMessage = nil
	if !processEntityInteract(s, rp, 13) || chest.State != "shut" {
		t.Fatalf("interacting with the open chest should close it, got %s", chest.State)
	}
	if !processEntityInteract(s, rp, 13) || chest.State != "open" {
		t.Fatalf("interacting with the shut chest should reopen it, got %s", chest.State)
	}
	if rp.inventoryCount("note") != 1 || rp.PendingHudMessage != nil {
		t.Fatalf("reopening must not give the loot again (note=%d, msg=%v)", rp.inventoryCount("note"), rp.PendingHudMessage)
	}
	processEntityInteract(s, rp, 13)
	if chest.State != "shut" {
		t.Fatalf("and close again, got %s", chest.State)
	}
}

// Without interact.close, an opened entity stays open.
func TestOpenOnceWithoutClose(t *testing.T) {
	s, rp, _ := newPuzzleServer(t)
	s.entityDefinitions["crate"] = &entdefs.EntityDefinition{Title: "crate", Solid: true, InitialState: "closed", Interact: &entdefs.Interact{Open: "open"}}
	s.entityInstances[20] = &Entity{ServerId: 20, Definition: "crate", State: "closed", X: 48, WorldId: "testWorld", GridId: "testGrid"}
	rp.Entity.X = 48
	processEntityInteract(s, rp, 20)
	if processEntityInteract(s, rp, 20) || s.entityInstances[20].State != "open" {
		t.Fatal("an open-once entity has nothing more to do")
	}
}

// The real chest closes into an animated "shut" state.
func TestChestDefinitionCloses(t *testing.T) {
	def, err := entdefs.ReadEntdefFile("../../dat/entdefs/chest.yml")
	if err != nil {
		t.Fatal(err)
	}
	if def.Interact == nil || def.Interact.Open != "open" || def.Interact.Close != "shut" {
		t.Fatalf("chest interact: %+v", def.Interact)
	}
	cat, err := tiled.LoadEntityVisualCatalog("../../dat/entdefs")
	if err != nil {
		t.Fatal(err)
	}
	vis, ok := cat.Lookup("chest_shut")
	if !ok || len(vis.Frames) != 4 || vis.Frames[3] != 10 || vis.TileProps["loop"] != "false" {
		t.Fatalf("chest_shut should be a play-once animation, got %+v (found %v)", vis, ok)
	}
}
