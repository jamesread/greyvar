package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/gridfiles"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/server/pkg/worlds"
)

const mechWorld, mechGrid = "mechWorld", "0.0.tmj"

// mechServer is the real server data (entdefs, tilesets, player hitbox) plus
// an 8x8 floor whose column 7 is wall, for push/plate/timer tests.
func mechServer(t *testing.T) *serverInterface {
	t.Helper()
	s := tailorServer(t)
	saved := blockingTileTextures
	t.Cleanup(func() { blockingTileTextures = saved })
	blockingTileTextures = map[string]bool{"wall.png": true}
	grid := &gridfiles.Grid{RowCount: 8, ColCount: 8}
	grid.Build()
	for r := uint32(0); r < 8; r++ {
		for c := uint32(0); c < 8; c++ {
			tex := "floor.png"
			if c == 7 {
				tex = "wall.png"
			}
			grid.Tiles[r][c] = &gridfiles.Tile{Row: r, Col: c, Texture: tex}
		}
	}
	s.loadedWorlds[mechWorld] = &worlds.World{ID: mechWorld, Grids: map[string]*gridfiles.Grid{mechGrid: grid}}
	return s
}

func mechEnt(s *serverInterface, def string, row, col int32, props map[string]string) *Entity {
	s.loadEntdef(def)
	e := &Entity{ServerId: s.nextEntityId(), Definition: def, State: s.entityDefinitions[def].InitialState,
		X: col * 16, Y: row * 16, WorldId: mechWorld, GridId: mechGrid, Properties: props}
	s.entityInstances[e.ServerId] = e
	return e
}

func mechPlayer(s *serverInterface, name string, x, y int32) *RemotePlayer {
	ent := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: x, Y: y, WorldId: mechWorld, GridId: mechGrid}
	s.entityInstances[ent.ServerId] = ent
	rp := &RemotePlayer{Username: name, Entity: ent, CurrentWorldId: mechWorld, CurrentGridId: mechGrid,
		KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	for _, e := range s.entitiesOnGrid(mechWorld, mechGrid) {
		rp.KnownEntities[e.ServerId] = e
	}
	if s.remotePlayers == nil || s.remotePlayers["visitor"] != nil {
		s.remotePlayers = map[string]*RemotePlayer{}
	}
	s.remotePlayers[name] = rp
	return rp
}

func walk(s *serverInterface, rp *RemotePlayer, dx, dy int32, moves int) {
	for i := 0; i < moves; i++ {
		processMoveRequest(s, rp, &pb.MoveRequest{X: dx, Y: dy})
	}
}

// standLeftOf puts the player flush against the left of tile (row, col).
func standLeftOf(rp *RemotePlayer, row, col int32) {
	hb := playerHitbox
	rp.Entity.X = col*16 - hb.X - hb.W
	rp.Entity.Y = row*16 + 8 - hb.Y - hb.H/2
}

func TestPushBlockSlidesOneTileAtATime(t *testing.T) {
	s := mechServer(t)
	block := mechEnt(s, "pushBlock", 3, 3, nil)
	rp := mechPlayer(s, "p1", 0, 0)
	standLeftOf(rp, 3, 3)

	// A brush against it does nothing; leaning on it pushes it one tile.
	walk(s, rp, 1, 0, pushHoldMoves-1)
	if block.X != 48 {
		t.Fatalf("block moved after %d moves: x=%d", pushHoldMoves-1, block.X)
	}
	walk(s, rp, 1, 0, 1)
	if block.X != 64 || block.Y != 48 {
		t.Fatalf("block should slide one tile right, at %d,%d", block.X, block.Y)
	}

	// Keep going: it slides to col 6 and stops at the wall in col 7.
	walk(s, rp, 1, 0, 60)
	if block.X != 96 {
		t.Fatalf("block should stop against the wall at col 6, x=%d", block.X)
	}
	if blocked, _ := s.isBlockedBySolidEntity(rp, rp.Entity, block.X-playerHitbox.X-playerHitbox.W+1, rp.Entity.Y); !blocked {
		t.Fatal("the block should still be solid")
	}
}

func TestPushBlockNeedsSquareContact(t *testing.T) {
	s := mechServer(t)
	block := mechEnt(s, "pushBlock", 3, 3, nil)
	rp := mechPlayer(s, "p1", 0, 0)
	standLeftOf(rp, 3, 3)
	// Mostly below the block: walking right slides past its corner, no push.
	rp.Entity.Y += 13
	walk(s, rp, 1, 0, 3*pushHoldMoves)
	if block.X != 48 {
		t.Fatalf("an off-centre brush pushed the block to x=%d", block.X)
	}
}

func TestPushBlockStopsAtBlocksDoorsAndPlayers(t *testing.T) {
	s := mechServer(t)
	block := mechEnt(s, "pushBlock", 1, 2, nil)
	mechEnt(s, "pushBlock", 1, 3, nil)
	rp := mechPlayer(s, "p1", 0, 0)
	standLeftOf(rp, 1, 2)
	walk(s, rp, 1, 0, 3*pushHoldMoves)
	if block.X != 32 {
		t.Fatal("a block must not push another block")
	}

	block2 := mechEnt(s, "pushBlock", 4, 2, nil)
	door := mechEnt(s, "cellDoor", 4, 3, map[string]string{"open_when": "nothing=here"})
	s.evaluateOpenWhen(mechWorld, mechGrid, 0)
	if door.State != "closed" {
		t.Fatalf("door state %q", door.State)
	}
	standLeftOf(rp, 4, 2)
	walk(s, rp, 1, 0, 3*pushHoldMoves)
	if block2.X != 32 {
		t.Fatal("a block must not go into a doorway")
	}

	block3 := mechEnt(s, "pushBlock", 6, 2, nil)
	other := mechPlayer(s, "p2", 3*16, 6*16)
	standLeftOf(rp, 6, 2)
	walk(s, rp, 1, 0, 3*pushHoldMoves)
	if block3.X != 32 {
		t.Fatalf("a block must not be pushed onto a player (p2 at %d,%d)", other.Entity.X, other.Entity.Y)
	}
}

func TestBlockHoldsPlateDown(t *testing.T) {
	s := mechServer(t)
	block := mechEnt(s, "pushBlock", 2, 2, nil)
	plate := mechEnt(s, "pressureButton", 2, 3, map[string]string{"id": "plate1"})
	door := mechEnt(s, "cellDoor", 5, 5, map[string]string{"open_when": "plate1=pressed"})
	s.evaluateOpenWhen(mechWorld, mechGrid, 0)
	rp := mechPlayer(s, "p1", 0, 0)
	standLeftOf(rp, 2, 2)

	walk(s, rp, 1, 0, pushHoldMoves)
	if block.X != 48 || plate.State != "pressed" || door.State != "open" {
		t.Fatalf("block on the plate: block x=%d plate %q door %q", block.X, plate.State, door.State)
	}
	// Push it off the far side: the plate pops up and the door shuts.
	walk(s, rp, 1, 0, 40)
	if block.X <= 48 || plate.State != "unpressed" || door.State != "closed" {
		t.Fatalf("block off the plate: block x=%d plate %q door %q", block.X, plate.State, door.State)
	}
	// Walking onto it still presses it (players count too).
	rp.Entity.X, rp.Entity.Y = plate.X, plate.Y
	s.processEntityInteractions(rp)
	if plate.State != "pressed" {
		t.Fatal("a player standing on the plate should press it")
	}
}

func TestPlateStaysDownWhileAnyoneIsOnIt(t *testing.T) {
	s := mechServer(t)
	plate := mechEnt(s, "pressureButton", 2, 2, map[string]string{"id": "plate1"})
	a := mechPlayer(s, "p1", 32, 32)
	b := mechPlayer(s, "p2", 32, 32)
	s.processEntityInteractions(a)
	s.processEntityInteractions(b)
	if plate.State != "pressed" {
		t.Fatal("plate should be pressed")
	}
	a.Entity.X = 80
	s.processEntityInteractions(a)
	if plate.State != "pressed" {
		t.Fatal("p2 is still on the plate; it must stay down")
	}
	// p2 leaves the grid altogether: nobody is left on it.
	s.leaveCurrentLocation(b)
	if plate.State != "unpressed" {
		t.Fatalf("plate should pop up once everyone has gone, state %q", plate.State)
	}
}

func TestTimedLeverSpringsBack(t *testing.T) {
	s := mechServer(t)
	lever := mechEnt(s, "wallSwitch", 2, 2, map[string]string{"id": "t1", "reset_after": "1500"})
	door := mechEnt(s, "cellDoor", 5, 5, map[string]string{"open_when": "t1=flickedRight"})
	s.evaluateOpenWhen(mechWorld, mechGrid, 0)
	rp := mechPlayer(s, "p1", 2*16, 3*16)

	if !processEntityInteract(s, rp, lever.ServerId) || lever.State != "flickedRight" || door.State != "open" {
		t.Fatalf("flip: lever %q door %q", lever.State, door.State)
	}
	if lever.resetAt == 0 || rp.PendingHudMessage == nil || *rp.PendingHudMessage == "" {
		t.Fatal("a timed lever should start counting down and say so")
	}
	at := lever.resetAt
	s.tickTimers(at - 1)
	if lever.State != "flickedRight" || door.State != "open" {
		t.Fatal("lever sprang back early")
	}
	s.tickTimers(at)
	if lever.State != "flickedLeft" || door.State != "closed" || lever.resetAt != 0 {
		t.Fatalf("after the timer: lever %q door %q", lever.State, door.State)
	}

	// Flipping it back by hand before time stops the countdown.
	processEntityInteract(s, rp, lever.ServerId)
	processEntityInteract(s, rp, lever.ServerId)
	if lever.State != "flickedLeft" || lever.resetAt != 0 {
		t.Fatalf("hand reset: lever %q resetAt %d", lever.State, lever.resetAt)
	}

	// Without reset_after a lever stays where it is put.
	plain := mechEnt(s, "wallSwitch", 2, 4, map[string]string{"id": "p"})
	rp.Entity.X = 4 * 16
	processEntityInteract(s, rp, plain.ServerId)
	if plain.State != "flickedRight" || plain.resetAt != 0 {
		t.Fatal("an ordinary lever must not reset")
	}
}

// standInCell puts the player squarely in tile (row, col).
func standInCell(rp *RemotePlayer, row, col int32) {
	rp.Entity.X, rp.Entity.Y = col*16, row*16
}

func TestPullBlockOutOfACorner(t *testing.T) {
	s := mechServer(t)
	// Wedged against the wall (col 7) in the top-right corner: pushing can
	// never move it again, pulling can.
	block := mechEnt(s, "pushBlock", 0, 6, nil)
	rp := mechPlayer(s, "p1", 0, 0)
	standInCell(rp, 0, 5)
	if !processEntityInteract(s, rp, block.ServerId) {
		t.Fatal("use on a block beside you should pull it")
	}
	if block.X != 5*16 || block.Y != 0 || rp.Entity.X != 4*16 || rp.Entity.Y != 0 {
		t.Fatalf("after pulling left: block %d,%d player %d,%d", block.X, block.Y, rp.Entity.X, rp.Entity.Y)
	}
	// From below, pulling down: out of the top row.
	standInCell(rp, 1, 5)
	if !processEntityInteract(s, rp, block.ServerId) || block.Y != 16 || rp.Entity.Y != 32 {
		t.Fatalf("after pulling down: block %d,%d player %d,%d", block.X, block.Y, rp.Entity.X, rp.Entity.Y)
	}
}

func TestPullNeedsRoomBehindYou(t *testing.T) {
	s := mechServer(t)
	block := mechEnt(s, "pushBlock", 3, 2, nil)
	mechEnt(s, "pushBlock", 3, 0, nil) // right behind the player
	rp := mechPlayer(s, "p1", 0, 0)
	standInCell(rp, 3, 1)
	if processEntityInteract(s, rp, block.ServerId) || block.X != 32 || rp.Entity.X != 16 {
		t.Fatalf("no room to step back: block x=%d player x=%d", block.X, rp.Entity.X)
	}
	// At the grid's edge there is nowhere to step back to either.
	edge := mechEnt(s, "pushBlock", 5, 1, nil)
	standInCell(rp, 5, 0)
	if processEntityInteract(s, rp, edge.ServerId) || edge.X != 16 || rp.Entity.X != 0 {
		t.Fatal("pulling at the edge of the map must not move anything")
	}
}

func TestPullNotFromDiagonal(t *testing.T) {
	s := mechServer(t)
	block := mechEnt(s, "pushBlock", 3, 3, nil)
	rp := mechPlayer(s, "p1", 0, 0)
	standInCell(rp, 4, 4) // diagonally below-right
	if processEntityInteract(s, rp, block.ServerId) || block.X != 48 || block.Y != 48 {
		t.Fatalf("diagonal pull moved the block to %d,%d", block.X, block.Y)
	}
}

func TestPulledBlockPressesPlate(t *testing.T) {
	s := mechServer(t)
	block := mechEnt(s, "pushBlock", 2, 3, nil)
	plate := mechEnt(s, "pressureButton", 2, 2, map[string]string{"id": "plate1"})
	rp := mechPlayer(s, "p1", 0, 0)
	standInCell(rp, 2, 2) // standing on the plate, pulling the block onto it
	s.processEntityInteractions(rp)
	if !processEntityInteract(s, rp, block.ServerId) {
		t.Fatal("pull failed")
	}
	if block.X != 32 || plate.State != "pressed" {
		t.Fatalf("block x=%d plate %q", block.X, plate.State)
	}
}
