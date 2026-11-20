package greyvarserver

import (
	"github.com/coder/websocket"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

type InventoryItem struct {
	Definition string
	Count      int32
}

type RemotePlayer struct {
	// sentPositions: entity positions this player was last sent (frameEntityPositions).
	sentPositions     map[int64][2]int32
	sentPositionsGrid string

	// dialog is the NPC conversation this player is in (dialog.go); nil if none.
	dialog *activeDialog

	Connection      *websocket.Conn
	Username        string
	// key identifies this connection in serverInterface.remotePlayers.
	key string
	NeedsGridUpdate bool
	Spawned         bool

	Entity *Entity

	CurrentGridId  string
	CurrentWorldId string

	PendingGridTransition *GridTransitionInfo

	KnownEntities          map[int64]*Entity
	KnownEntdefs           map[string]bool
	KnownItemdefs          map[string]bool
	KnownRecipes           map[string]bool
	PendingDespawns        []int64
	PendingConsoleMessages []string
	PendingAppearanceChanges []*pb.EntityAppearanceChange
	PendingTileChanges       []*pb.TileChange

	Inventory              []InventoryItem
	PendingInventoryUpdate bool

	// Touch HUD / contact tracking.
	ActiveSignEntityId int64
	ActiveSignMsg      string
	PendingHudMessage  *string
	PendingHudEntityId int64
	// pendingHudSticky is the sign message that stays until replaced or
	// cleared; any other PendingHudMessage (set after it) fades on the client.
	pendingHudSticky *string
	ActiveEntityIds    map[int64]bool

	// Pushing a block: consecutive blocked moves into the same pushable
	// entity in the same direction (see tryPush).
	pushTarget int64
	pushDirX   int32
	pushDirY   int32
	pushCount  int

	TimeOfLastMoveRequest int64

	currentFrame *pb.ServerUpdate

	pendingRequests []*pb.ClientRequests
}

func (rp *RemotePlayer) addInventoryItem(definition string, count int32) {
	if rp == nil || definition == "" || count == 0 {
		return
	}
	for i := range rp.Inventory {
		if rp.Inventory[i].Definition == definition {
			rp.Inventory[i].Count += count
			return
		}
	}
	rp.Inventory = append(rp.Inventory, InventoryItem{Definition: definition, Count: count})
}
