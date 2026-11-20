package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/server/pkg/worlds"
)

// 2.1 (south of 1.1, the spawn grid) is a sandy cove: 1.1's beach runs off
// its bottom edge, so walking south must lead somewhere, and the sea must
// stop you at the cove's far edges.
func TestSouthCoveJoinsUpWith1_1(t *testing.T) {
	s := tailorServer(t)
	const world = "isleOfStarting_dev_tiled"
	w := s.loadedWorlds[world]
	if dest, ok := worlds.AdjacentGridId(w, "1.1.tmj", 1, 0); !ok || dest != "2.1.tmj" {
		t.Fatalf("south of 1.1 should be 2.1, got %q %v", dest, ok)
	}

	// 1.1's sand reaches its bottom edge at cols 2-8; 2.1 carries it on.
	for col := int32(3); col <= 8; col++ {
		if !s.isTraversableOnGrid(world, "1.1.tmj", col*16, 15*16) {
			t.Errorf("1.1 bottom row col %d should be walkable", col)
		}
		if !s.isTraversableOnGrid(world, "2.1.tmj", col*16, 0) {
			t.Errorf("2.1 top row col %d should be walkable (continues 1.1's beach)", col)
		}
	}

	type pt struct{ x, y int32 }
	start := pt{5 * 16, 0}
	seen := map[pt]bool{start: true}
	queue := []pt{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range []pt{{4, 0}, {-4, 0}, {0, 4}, {0, -4}} {
			n := pt{p.x + d.x, p.y + d.y}
			if n.x < 0 || n.y < 0 || n.x > 240 || n.y > 240 || seen[n] || !s.isTraversableOnGrid(world, "2.1.tmj", n.x, n.y) {
				continue
			}
			seen[n] = true
			queue = append(queue, n)
		}
	}
	for p := range seen {
		if p.x == 0 || p.y == 240 {
			t.Fatalf("walked to the open sea at (%d,%d): the cove should be surrounded by water", p.x, p.y)
		}
	}
	if !seen[pt{8 * 16, 6 * 16}] {
		t.Error("the grassy dune in the middle of the cove should be reachable")
	}
	if !seen[pt{8 * 16, 12 * 16}] {
		t.Error("the southern tip of the cove should be reachable")
	}

	var sign *Entity
	for _, e := range s.entitiesOnGrid(world, "2.1.tmj") {
		if e.Definition == "signpost" {
			sign = e
		}
	}
	if sign == nil {
		t.Fatal("2.1 should have a signpost")
	}
}
