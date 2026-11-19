import Tile from './Tile.js'

export default class Grid {
  constructor(rowCount, colCount) {
    this.tileMap = []
    this.rowCount = rowCount
    this.colCount = colCount
    this.gridId = null
    this.worldId = null
    this.tilesets = []
    this.transition = null
    this.layers = []
  }

  rebuildTopTileMap () {
    this.tileMap = []

    for (let row = 0; row < this.rowCount; row++) {
      this.tileMap[row] = []
      for (let col = 0; col < this.colCount; col++) {
        const top = this.topTileAt(row, col)
        this.tileMap[row][col] = top ?? new Tile()
      }
    }
  }

  topTileAt (row, col) {
    for (let i = this.layers.length - 1; i >= 0; i--) {
      const layer = this.layers[i]
      for (const tile of layer.tiles) {
        if (tile.row === row && tile.col === col) {
          return tile
        }
      }
    }
    return null
  }

  /**
   * The tile whose collision applies here, as the server decides it
   * (gridfiles.Grid.CollisionTileAt): the topmost tile that is neither decor
   * nor an overlay, with every overlay's rects above it added. Null if empty.
   */
  collisionTileAt (row, col) {
    const extra = []
    let base = null
    for (let i = this.layers.length - 1; i >= 0 && base == null; i--) {
      const tile = this.layers[i].tiles.find((t) => t.row === row && t.col === col)
      if (tile == null || tile.decor) {
        continue
      }
      if (tile.overlay) {
        extra.push(...(tile.collision ?? []))
        continue
      }
      base = tile
    }
    if (extra.length === 0) {
      return base
    }
    const merged = Object.assign(Object.create(Object.getPrototypeOf(base ?? {})), base ?? { row, col, collisionFromTsx: true })
    merged.collision = [...(base?.collision ?? []), ...extra]
    return merged
  }

  allLayerTiles () {
    return this.layers.flatMap((layer) => layer.tiles)
  }

  allTiles() {
    if (this.layers.length > 0) {
      return this.tileMap.flat()
    }

    return this.tileMap.flat()
  }

  set (row, col, tile) {
    if (this.tileMap[row] == undefined) {
      console.error('grid set out of bounds', { row, col, rowCount: this.rowCount, colCount: this.colCount })
      return
    }

    this.tileMap[row][col] = tile
  }
}
