package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func TestUpdateSignHudEnterAndLeave(t *testing.T) {
	grid := &gridfiles.Grid{RowCount: 4, ColCount: 4}
	grid.Build()
	s, rp := testServerWithGrid(grid)
	s.entityDefinitions["signpost"] = &entdefs.EntityDefinition{
		Title:        "signpost",
		InitialState: "signy",
		States:       map[string]entdefs.EntityState{"signy": {}},
	}
	s.entityVisuals = &tiled.EntityVisualCatalog{}
	s.entityVisuals.SetTypeForTest("signpost", tiled.TypedTileVisual{
		Frames:        []int32{13},
		Collision:     []tiled.CollisionRect{{X: 0, Y: 0, W: 15, H: 15}},
		CollisionType: tiled.CollisionTouch,
		TileWidth:     15,
		TileHeight:    15,
	})

	player := &Entity{
		ServerId: 1, Definition: "player",
		X: 15, Y: 0, WorldId: "testWorld", GridId: "testGrid",
	}
	sign := &Entity{
		ServerId: 2, Definition: "signpost", State: "signy",
		Properties: map[string]string{"msg": "Hello there!"},
		X: 32, Y: 0, WorldId: "testWorld", GridId: "testGrid",
	}
	s.entityInstances[player.ServerId] = player
	s.entityInstances[sign.ServerId] = sign
	rp.Entity = player
	rp.currentFrame = &pb.ServerUpdate{}

	s.processEntityInteractions(rp)
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage != "Hello there!" {
		t.Fatalf("pending hud = %#v, want Hello there!", rp.PendingHudMessage)
	}

	rp.PendingHudMessage = nil
	s.processEntityInteractions(rp)
	if rp.PendingHudMessage != nil {
		t.Fatal("expected no hud update while still near same sign")
	}

	player.X = 64
	player.Y = 64
	s.processEntityInteractions(rp)
	if rp.PendingHudMessage == nil || *rp.PendingHudMessage != "" {
		t.Fatalf("pending hud = %#v, want clear", rp.PendingHudMessage)
	}
}

func TestPlayerNearEntityCollisionMargin(t *testing.T) {
	rects := []tiled.CollisionRect{{X: 0, Y: 0, W: 15, H: 15}}

	distTouch := aabbDistanceToCollisionRects(0, 0, 16, 16, rects, 16, 0)
	if distTouch != 0 {
		t.Fatalf("touching distance = %v, want 0", distTouch)
	}

	distGap := aabbDistanceToCollisionRects(-4, 0, 16, 16, rects, 16, 0)
	if distGap != 4 {
		t.Fatalf("gap distance = %v, want 4", distGap)
	}

	distClose := aabbDistanceToCollisionRects(-2, 0, 16, 16, rects, 16, 0)
	if distClose != 2 {
		t.Fatalf("close distance = %v, want 2", distClose)
	}
	if distClose > float64(entityTriggerMarginPx) {
		t.Fatal("gap 2 should be within margin 3")
	}
}

// Sign text stays (sticky) while you stand there; other messages, even one
// set after a sign's in the same frame, fade on the client.
func TestHudMessageStickyOnlyForSigns(t *testing.T) {
	rp := &RemotePlayer{currentFrame: &pb.ServerUpdate{}}
	sign := "Welcome!"
	rp.PendingHudMessage, rp.pendingHudSticky = &sign, &sign
	frameHudMessages(nil, rp)
	if !rp.currentFrame.HudMessage.Sticky {
		t.Fatal("a sign's message should be sticky")
	}

	rp.currentFrame = &pb.ServerUpdate{}
	rp.PendingHudMessage, rp.pendingHudSticky = &sign, &sign
	got := "Got Hay x1"
	rp.PendingHudMessage = &got // e.g. a harvest later in the same frame
	frameHudMessages(nil, rp)
	if rp.currentFrame.HudMessage.Sticky || rp.currentFrame.HudMessage.Text != got {
		t.Fatalf("a replacing pickup message must fade, got %+v", rp.currentFrame.HudMessage)
	}
	if rp.pendingHudSticky != nil {
		t.Fatal("sticky marker should reset after sending")
	}
}
