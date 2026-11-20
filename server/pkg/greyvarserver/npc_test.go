package greyvarserver

import (
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func billServer(t *testing.T) (*serverInterface, *Entity) {
	t.Helper()
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	npcRand = rand.New(rand.NewSource(1))
	saved := playerHitbox // newServer sets it from the rig; other tests expect the default
	t.Cleanup(func() { playerHitbox = saved })
	s := newServer()
	for _, e := range s.entitiesOnGrid("isleOfStarting_dev_tiled", "1.2.redHouse.tmj") {
		if e.Definition == "bill" {
			return s, e
		}
	}
	t.Fatal("Bill is not in the red house (isleOfStarting 1.2.redHouse)")
	return nil, nil
}

func playerOn(s *serverInterface, grid string, x, y int32) *RemotePlayer {
	ent := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: x, Y: y, WorldId: "isleOfStarting_dev_tiled", GridId: grid}
	s.entityInstances[ent.ServerId] = ent
	rp := &RemotePlayer{Username: "visitor", Entity: ent, CurrentWorldId: "isleOfStarting_dev_tiled", CurrentGridId: grid,
		KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	for _, e := range s.entitiesOnGrid(rp.CurrentWorldId, grid) {
		rp.KnownEntities[e.ServerId] = e
	}
	s.remotePlayers = map[string]*RemotePlayer{"visitor": rp}
	return rp
}

func TestBillLooksTheSameEveryTime(t *testing.T) {
	_, bill := billServer(t)
	a := bill.Appearance
	if a == nil || a.Palettes["shirt"] != "green" || a.Palettes["hair"] != "grey" || a.Palettes["skin"] != "tan" || a.Parts["hair"] != "short" || a.Parts["beard"] != "beard" || a.Rig != "adult" {
		t.Fatalf("Bill's fixed look: %+v", a)
	}
}

func TestBillWandersNearHomeOnlyWhenSeen(t *testing.T) {
	s, bill := billServer(t)
	homeX, homeY := bill.X, bill.Y
	now := nowMs()

	s.remotePlayers = map[string]*RemotePlayer{}
	for i := 0; i < 200; i++ {
		s.tickNPCs(now + int64(i)*100)
	}
	if bill.X != homeX {
		t.Fatal("with nobody on his grid, Bill should not be ticked")
	}

	playerOn(s, "1.2.redHouse.tmj", 0, 0) // in the corner of the map, away from him
	minX, maxX := bill.X, bill.X
	for i := 0; i < 3000; i++ {
		s.tickNPCs(now + int64(i)*100)
		if bill.X < minX {
			minX = bill.X
		}
		if bill.X > maxX {
			maxX = bill.X
		}
		if bill.Y != homeY {
			t.Fatalf("Bill only walks side to side, but moved to y=%d", bill.Y)
		}
	}
	if minX < homeX-24 || maxX > homeX+24 {
		t.Fatalf("Bill strayed: x %d..%d, home %d", minX, maxX, homeX)
	}
	if maxX-minX < 12 {
		t.Fatalf("Bill should stroll a little, only moved x %d..%d", minX, maxX)
	}
}

func TestBillDoesNotWalkIntoPlayers(t *testing.T) {
	s, bill := billServer(t)
	rp := playerOn(s, "1.2.redHouse.tmj", bill.X+12, bill.Y)
	if s.npcCanStand(bill, bill.X+9, bill.Y) {
		t.Fatal("Bill should not step into a player")
	}
	_ = rp
}

func TestTalkingToBillGivesSeeds(t *testing.T) {
	s, bill := billServer(t)
	rp := playerOn(s, "1.2.redHouse.tmj", bill.X+16, bill.Y)
	s.loadItemCatalog()

	if !processEntityInteract(s, rp, bill.ServerId) || rp.currentFrame.Dialog == nil {
		t.Fatal("interacting with Bill should open his dialog")
	}
	d := rp.currentFrame.Dialog
	if d.Speaker != "Bill" || d.Node != "hello" || len(d.Options) != 3 || !d.Options[2].Ends {
		t.Fatalf("Bill's greeting: %+v", d)
	}
	if bill.npc.walking || bill.npc.waitUntil <= nowMs() {
		t.Fatal("Bill should stop to talk")
	}
	var faced bool
	for _, c := range rp.currentFrame.EntityStateChanges {
		faced = faced || (c.EntityId == bill.ServerId && c.NewState == "faceRight")
	}
	if !faced {
		t.Fatal("Bill should turn to face the player (to his right)")
	}

	// A stale or made-up choice does nothing.
	rp.currentFrame = &pb.ServerUpdate{}
	processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: bill.ServerId, Node: "seeds", Option: 0})
	processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: bill.ServerId, Node: "hello", Option: 7})
	if rp.currentFrame.Dialog != nil || rp.inventoryCount("hay_seeds") != 0 {
		t.Fatal("only valid choices for the shown node count")
	}

	processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: bill.ServerId, Node: "hello", Option: 0})
	d = rp.currentFrame.Dialog
	if d == nil || d.Node != "seeds" || rp.inventoryCount("hay_seeds") != 5 || !strings.Contains(d.Text, "Got") {
		t.Fatalf("asking for seeds should give 5 hay seeds: dialog %+v, seeds %d", d, rp.inventoryCount("hay_seeds"))
	}
	processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: bill.ServerId, Node: "seeds", Option: 0})
	if rp.dialog != nil {
		t.Fatal("'Thanks, Bill!' ends the conversation")
	}

	// Walk away mid-conversation: choices no longer count.
	processEntityInteract(s, rp, bill.ServerId)
	rp.Entity.X = bill.X + 200
	processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: bill.ServerId, Node: "hello", Option: 0})
	if rp.inventoryCount("hay_seeds") != 5 {
		t.Fatal("out of range: no more seeds")
	}
}
