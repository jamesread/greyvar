package greyvarserver

import (
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// isleOfStarting 1.0: the hill's cliffs block, and the steps are the way up.
func TestHillStepsAreTheWayUp(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	saved := playerHitbox
	defer func() { playerHitbox = saved }()

	s := newServer()
	const world, grid = "isleOfStarting_dev_tiled", "1.0.tmj"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", WorldId: world, GridId: grid}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "climber", Entity: player, CurrentWorldId: world, CurrentGridId: grid,
		KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"climber": rp}
	g := s.gridById(world, grid)

	// blocked lets the search treat chosen cells as solid (to prove the
	// steps are the only way up).
	blocked := map[[2]int32]bool{}
	reachable := func(fromRow, fromCol, toRow, toCol int32) bool {
		type pos struct{ x, y int32 }
		start, goal := pos{fromCol * 16, fromRow * 16}, pos{toCol * 16, toRow * 16}
		seen := map[pos]bool{start: true}
		queue := []pos{start}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			if cur == goal {
				return true
			}
			for _, d := range []pos{{2, 0}, {-2, 0}, {0, 2}, {0, -2}} {
				n := pos{cur.x + d.x, cur.y + d.y}
				if seen[n] || !s.isTraversableAt(rp, n.x, n.y) {
					continue
				}
				hb := playerHitbox
				r0, r1 := (n.y+hb.Y)/16, (n.y+hb.Y+hb.H-1)/16
				c0, c1 := (n.x+hb.X)/16, (n.x+hb.X+hb.W-1)/16
				solid := false
				for r := r0; r <= r1; r++ {
					for c := c0; c <= c1; c++ {
						solid = solid || blocked[[2]int32{r, c}]
					}
				}
				if solid {
					continue
				}
				seen[n] = true
				queue = append(queue, n)
			}
		}
		return false
	}

	if tile := g.CollisionTileAt(6, 13); tile.AtlasKey != "steps" || len(tile.Collision) != 0 {
		t.Fatalf("(6,13) should be walkable steps, got %s %+v", tile.AtlasKey, tile.Collision)
	}
	for _, rc := range [][2]uint32{{6, 9}, {6, 10}, {6, 14}, {4, 4}} {
		if tile := g.CollisionTileAt(rc[0], rc[1]); len(tile.Collision) == 0 {
			t.Errorf("cliff/corner at %v should have collision", rc)
		}
	}

	// The taller face (res/sprites/faces) hangs 12px into the sand below the
	// cliff: blocked at its top, walkable beneath it.
	if s.isTraversableAt(rp, 11*16, 7*16) {
		t.Error("the face extension under the cliff at (6,11) should block the top of (7,11)")
	}
	if !s.isTraversableAt(rp, 11*16, 8*16) {
		t.Error("the sand a tile further down should stay walkable")
	}

	// Beach (7,13) -> plateau (5,13) via the steps; dip (3,7) -> plateau (1,7).
	if !reachable(7, 13, 5, 13) {
		t.Error("the beach steps at (6,13) should lead up onto the hill")
	}
	if !reachable(3, 7, 1, 7) {
		t.Error("the dip steps at (2,7) should lead up onto the hill")
	}
	// Without the steps there is no way up: the cliffs (and fixed corners) hold.
	blocked[[2]int32{6, 13}] = true
	blocked[[2]int32{2, 7}] = true
	if reachable(7, 13, 5, 13) {
		t.Error("with the steps closed, the hill must not be reachable from the beach")
	}
	if reachable(3, 7, 1, 7) {
		t.Error("with the steps closed, the hill must not be reachable from the dip")
	}
	// Between the taller cliff face and the water the beach stays wide enough
	// to walk along: the steps' beach (7,13) still reaches the west beach.
	if !reachable(7, 13, 7, 4) {
		t.Error("the beach under the cliff should connect the steps to the west beach")
	}
}
