package greyvarserver

import "github.com/jamesread/greyvar/datlib/rigs"

type Entity struct {
	ServerId int64

	Texture string
	Spawned bool

	Definition string
	State      string

	// Properties are Tiled instance custom properties (e.g. msg on signs).
	Properties map[string]string

	X int32
	Y int32

	GridId  string
	WorldId string

	ServerDebugAlias string

	// Appearance is the layered-rig look (players and NPCs); nil otherwise.
	Appearance *rigs.Appearance

	// npc is the wander state of an NPC (entdef wander); nil otherwise.
	npc *npcState

	// resetAt (unix ms) springs a reset_after lever back to its initial
	// state; 0 when not counting down.
	resetAt int64
}
