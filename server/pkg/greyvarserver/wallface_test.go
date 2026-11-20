package greyvarserver

import (
	"path/filepath"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

// isleOfStarting 2.1: the castle's south wall face continues a full row
// down (its lower half on "wall faces (generated)"), and that row blocks.
func TestTallWallFaceBlocks(t *testing.T) {
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")

	s := newServer()
	const world, grid = "isleOfStarting_dev_tiled", "2.1.tmj"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", WorldId: world, GridId: grid}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "visitor", Entity: player, CurrentWorldId: world, CurrentGridId: grid,
		KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"visitor": rp}

	// Feet at the bottom of a cell: hitbox bottom edge at (row+1)*16 - 1.
	standAt := func(row, col int32) bool {
		hb := playerHitbox
		return s.isTraversableAt(rp, col*16+8-hb.X-hb.W/2, (row+1)*16-hb.Y-hb.H)
	}
	if standAt(4, 12) {
		t.Error("row 4 under the castle wall is its lower face and should block")
	}
	if !standAt(5, 12) {
		t.Error("the grass below the face should be walkable")
	}
}
