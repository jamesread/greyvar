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
	Connection      *websocket.Conn
	Username        string
	NeedsGridUpdate bool
	Spawned         bool

	Entity *Entity

	CurrentGridId  string
	CurrentWorldId string

	PendingGridTransition *GridTransitionInfo

	KnownEntities          map[int64]*Entity
	KnownEntdefs           map[string]bool
	PendingDespawns        []int64
	PendingConsoleMessages []string

	Inventory              []InventoryItem
	PendingInventoryUpdate bool

	// Touch HUD / contact tracking.
	ActiveSignEntityId int64
	ActiveSignMsg      string
	PendingHudMessage  *string
	PendingHudEntityId int64
	ActiveEntityIds    map[int64]bool

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
