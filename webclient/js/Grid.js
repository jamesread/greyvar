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
