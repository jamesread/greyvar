package greyvarserver

import (
	"strings"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// Herman's Rift Bank on isleOfStarting 0.1: in through the door, talk to
// Herman across the counter, deposit rift coins, and they stay banked.
func TestHermanBanksRiftCoins(t *testing.T) {
	s := tailorServer(t)
	s.loadItemCatalog()
	const world = "isleOfStarting_dev_tiled"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 7 * 16, Y: 10 * 16, WorldId: world, GridId: "0.1.tmj"}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "saver", Entity: player, CurrentWorldId: world, CurrentGridId: "0.1.tmj",
		KnownEntities: map[int64]*Entity{}, KnownEntdefs: map[string]bool{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"saver": rp}
	walk := func(dy int32) {
		from := rp.CurrentGridId
		for i := 0; i < 12 && rp.CurrentGridId == from; i++ {
			s.frameTime += 70_000_000
			processMoveRequest(s, rp, &pb.MoveRequest{Y: dy})
		}
	}
	walk(-1)
	if rp.CurrentGridId != "0.1.hermanHouse.tmj" || player.X != 6*16 || player.Y != 10*16 {
		t.Fatalf("into the bank at (10,6), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}
	if s.isTraversableAt(rp, 6*16, 6*16) {
		t.Error("the teller counter should block the way behind it")
	}

	var herman *Entity
	for _, e := range s.entityInstances {
		if e.Definition == "herman" && e.GridId == "0.1.hermanHouse.tmj" {
			herman = e
		}
	}
	if herman == nil {
		t.Fatal("Herman is not in his bank")
	}
	player.X, player.Y = herman.X, herman.Y+3*16 // in front of the counter
	rp.addInventoryItem("rift_coin", 4)
	if !processEntityInteract(s, rp, herman.ServerId) || rp.currentFrame.Dialog == nil {
		t.Fatal("Herman should talk across the counter")
	}
	if d := rp.currentFrame.Dialog; d.Speaker != "Herman" || !strings.Contains(d.Text, "holds 0 rift coins") {
		t.Fatalf("greeting: %+v", d)
	}
	processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: herman.ServerId, Node: "hello", Option: 0})
	d := rp.currentFrame.Dialog
	if d == nil || d.Node != "deposited" || !strings.Contains(d.Text, "x4") || rp.inventoryCount("rift_coin") != 0 {
		t.Fatalf("deposit: %+v, carried %d", d, rp.inventoryCount("rift_coin"))
	}
	if got := s.accountBalance("saver", "rift_coin"); got != 4 {
		t.Fatalf("balance %d, want 4", got)
	}
	if got := (&serverInterface{}).accountBalance("saver", "rift_coin"); got != 4 {
		t.Fatalf("balance after a restart %d, want 4", got)
	}

	player.X, player.Y = 6*16, 10*16
	walk(+1)
	if rp.CurrentGridId != "0.1.tmj" || player.X != 7*16 || player.Y != 10*16 {
		t.Fatalf("back out in front of the bank (10,7), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}
}
