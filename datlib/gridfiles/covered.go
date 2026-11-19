package gridfiles

// Things drawn under a higher layer's tile are hidden, so they do not spawn:
// a flower (decor tile) under a hill or a sand edge, an entity on an object
// layer below a cliff's generated face. "Covered" means a later visible tile
// layer (terrain, overlays such as cliff faces) has a non-decor tile in the
// same cell. Ordinary terrain stacking is unaffected: only decor tiles,
// object-layer tiles and entities are dropped.

// dropCovered removes covered decor/object tiles and entities. entityLayer
// gives, per entity in g.Entities, the index in g.Layers of its object layer.
func dropCovered(g *Grid, entityLayer []int) (droppedTiles, droppedEntities int) {
	covered := func(layer int, row, col uint32) bool {
		for j := layer + 1; j < len(g.Layers); j++ {
			if g.Layers[j].Kind != LayerKindTerrain {
				continue
			}
			if t := g.Layers[j].TileAt(row, col); t != nil && !t.Decor {
				return true
			}
		}
		return false
	}

	for li := range g.Layers {
		l := &g.Layers[li]
		for row, cols := range l.Tiles {
			for col, t := range cols {
				if t == nil || (!t.Decor && l.Kind != LayerKindObject) {
					continue
				}
				if covered(li, row, col) {
					delete(cols, col)
					droppedTiles++
				}
			}
		}
	}

	kept := g.Entities[:0]
	for i, ent := range g.Entities {
		if i < len(entityLayer) && covered(entityLayer[i], ent.Row, ent.Col) {
			droppedEntities++
			continue
		}
		kept = append(kept, ent)
	}
	g.Entities = kept
	return droppedTiles, droppedEntities
}
