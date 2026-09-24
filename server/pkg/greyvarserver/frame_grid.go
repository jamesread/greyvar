package greyvarserver

import (
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/greyvar/datlib/gridfiles"
)

func frameGridUpdates(s *serverInterface, p *RemotePlayer) {
	if p.NeedsGridUpdate {
		p.currentFrame.Grid = generateGridUpdate(s, p)
		p.KnownEntities = make(map[int64]*Entity)
		// Keep KnownEntdefs across grids — definitions are world-global. Clearing
		// them forced a full resend and could race client sprite setup on return.
		p.NeedsGridUpdate = false
	}
}

func generateGridUpdate(s *serverInterface, p *RemotePlayer) (*pb.Grid) {
	worldId := p.CurrentWorldId
	gridId := p.CurrentGridId
	memGrid := s.gridById(worldId, gridId)
	if memGrid == nil {
		return nil
	}

	gridToSend := &pb.Grid{
		Title:    memGrid.Filename,
		GridId:   gridId,
		WorldId:  worldId,
		RowCount: memGrid.RowCount,
		ColCount: memGrid.ColCount,
	}

	for _, tileset := range memGrid.Tilesets {
		gridToSend.Tilesets = append(gridToSend.Tilesets, &pb.Tileset{
			Key:        tileset.Key,
			Image:      tileset.ImagePath,
			TileWidth:  uint32(tileset.TileWidth),
			TileHeight: uint32(tileset.TileHeight),
			Columns:    uint32(tileset.Columns),
		})
	}

	if p.PendingGridTransition != nil {
		gridToSend.Transition = &pb.GridTransition{
			FromGridId:   p.PendingGridTransition.FromGridId,
			ScrollDeltaX: p.PendingGridTransition.ScrollDeltaX,
			ScrollDeltaY: p.PendingGridTransition.ScrollDeltaY,
		}
		p.PendingGridTransition = nil
	}

	for _, layer := range memGrid.Layers {
		netLayer := &pb.GridLayer{
			Name: layer.Name,
			Kind: string(layer.Kind),
		}
		for row, cols := range layer.Tiles {
			for col, tile := range cols {
				_ = row
				_ = col
				if !tileVisibleToClient(tile) {
					continue
				}
				netLayer.Tiles = append(netLayer.Tiles, netTileFromMem(tile))
			}
		}
		gridToSend.Layers = append(gridToSend.Layers, netLayer)
	}

	for _, pos := range memGrid.CellIterator() {
		memTile := memGrid.TopTileAt(pos.Row, pos.Col)
		if memTile == nil || !tileVisibleToClient(memTile) {
			continue
		}
		gridToSend.Tiles = append(gridToSend.Tiles, netTileFromMem(memTile))
	}

	return gridToSend
}

func netTileFromMem(memTile *gridfiles.Tile) *pb.Tile {
	netTile := &pb.Tile{
		Row:              memTile.Row,
		Col:              memTile.Col,
		Tex:              memTile.Texture,
		Rot:              memTile.Rot,
		FlipH:            memTile.FlipH,
		FlipV:            memTile.FlipV,
		AtlasKey:         memTile.AtlasKey,
		FrameIndex:       memTile.FrameIndex,
		CollisionFromTsx: memTile.CollisionFromTSX,
	}

	for _, rect := range memTile.Collision {
		netTile.Collision = append(netTile.Collision, &pb.CollisionRect{
			X: float32(rect.X),
			Y: float32(rect.Y),
			W: float32(rect.W),
			H: float32(rect.H),
		})
	}

	return netTile
}

func tileVisibleToClient(memTile *gridfiles.Tile) bool {
	return memTile != nil && !memTile.Hidden
}
