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

	// Appearance is the layered-rig look (players only); nil otherwise.
	Appearance *rigs.Appearance
}
