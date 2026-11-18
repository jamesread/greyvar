package api

import "errors"

var (
	errInvalidEntdefName    = errors.New("invalid entdef name")
	errEntdefExists         = errors.New("entdef already exists")
	errTiledefNameRequired  = errors.New("tiledef name required")
	errUnknownTextureKind   = errors.New("unknown texture kind")
)
