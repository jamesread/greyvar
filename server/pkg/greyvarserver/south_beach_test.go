package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/server/pkg/worlds"
)

// 2.0 (south of 1.0) is the island's south beach: you walk in from 1.0's
// sand spit, can explore the beach, find its washed-up chest, and the sea
// stops you on every other side.
func TestSouthBeachJoinsUpWith1_0(t *testing.T) {
	s := tailorServer(t)
	const world = "isleOfStarting_dev_tiled"
	w := s.loadedWorlds[world]
	if dest, ok := worlds.AdjacentGridId(w, "1.0.tmj", 1, 0); !ok || dest != "2.0.tmj" {
		t.Fatalf("south of 1.0 should be 2.0, got %q %v", dest, ok)
	}

	// 1.0's beach reaches its bottom edge at cols 13-14 (a tide pool sits at
	// cols 11-12); 2.0's top row carries the sand on across cols 11-14.
	for col := int32(11); col <= 14; col++ {
		if col >= 13 && !s.isTraversableOnGrid(world, "1.0.tmj", col*16, 15*16) {
			t.Errorf("1.0 bottom row col %d should be walkable", col)
		}
		if !s.isTraversableOnGrid(world, "2.0.tmj", col*16, 0) {
			t.Errorf("2.0 top row col %d should be walkable (continues 1.0's beach)", col)
		}
	}

	// Explore from the entry in 4px steps.
	type pt struct{ x, y int32 }
	start := pt{13 * 16, 0}
	seen := map[pt]bool{start: true}
	queue := []pt{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range []pt{{4, 0}, {-4, 0}, {0, 4}, {0, -4}} {
			n := pt{p.x + d.x, p.y + d.y}
			if n.x < 0 || n.y < 0 || n.x > 240 || n.y > 240 || seen[n] || !s.isTraversableOnGrid(world, "2.0.tmj", n.x, n.y) {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}
	for p := range seen {
		if p.x == 0 || p.x == 240 || p.y == 240 {
			t.Fatalf("walked to the open sea at (%d,%d): the beach should be surrounded by water", p.x, p.y)
		}
	}
	if !seen[pt{8 * 16, 7 * 16}] {
		t.Error("the grassy dune in the middle of the beach should be reachable")
	}

	var chest, sign *Entity
	for _, e := range s.entitiesOnGrid(world, "2.0.tmj") {
		switch e.Definition {
		case "chest":
			chest = e
		case "signpost":
			sign = e
		}
	}
	if chest == nil || sign == nil {
		t.Fatal("2.0 should have a chest and a signpost")
	}
	near := false
	for p := range seen {
		if abs32(p.x-chest.X) <= 16 && abs32(p.y-chest.Y) <= 16 && (p.x != chest.X || p.y != chest.Y) {
			near = true
			break
		}
	}
	if !near {
		t.Fatalf("the chest at (%d,%d) should be reachable", chest.X, chest.Y)
	}

	rp := playerOn(s, "2.0.tmj", chest.X+16, chest.Y)
	s.loadItemCatalog()
	if !processEntityInteract(s, rp, chest.ServerId) || rp.inventoryCount("gem") != 1 {
		t.Fatalf("opening the beach chest should give a gem, have %d", rp.inventoryCount("gem"))
	}
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
