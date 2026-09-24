import Grid from './Grid.js'
import Tile from './Tile.js'

function tileFromNet (netTile) {
  const tile = new Tile()
  tile.fromNet(netTile)
  return tile
}

function buildLayersFromMessage (netGrid, rowCount, colCount) {
  const netLayers = netGrid.layers ?? []
  if (netLayers.length === 0) {
    return [{
      name: 'terrain',
      kind: 'terrain',
      tiles: (netGrid.tiles ?? []).map(tileFromNet)
    }]
  }

  return netLayers.map((layer) => ({
    name: layer.name ?? '',
    kind: layer.kind ?? 'terrain',
    tiles: (layer.tiles ?? []).map(tileFromNet)
  }))
}

export function buildGridFromMessage (netGrid) {
  if (netGrid == null) {
    return null
  }

  let rowCount = netGrid.rowCount ?? 0
  let colCount = netGrid.colCount ?? 0

  for (const netTile of netGrid.tiles ?? []) {
    rowCount = Math.max(rowCount, (netTile.row ?? 0) + 1)
    colCount = Math.max(colCount, (netTile.col ?? 0) + 1)
  }

  for (const layer of netGrid.layers ?? []) {
    for (const netTile of layer.tiles ?? []) {
      rowCount = Math.max(rowCount, (netTile.row ?? 0) + 1)
      colCount = Math.max(colCount, (netTile.col ?? 0) + 1)
    }
  }

  if (!rowCount || !colCount) {
    return null
  }

  const newGrid = new Grid(rowCount, colCount)
  newGrid.gridId = netGrid.gridId ?? netGrid.title ?? null
  newGrid.worldId = netGrid.worldId ?? null
  newGrid.tilesets = (netGrid.tilesets ?? []).map((tileset) => ({
    key: tileset.key,
    image: tileset.image,
    tileWidth: tileset.tileWidth ?? 16,
    tileHeight: tileset.tileHeight ?? 16,
    columns: tileset.columns ?? 0
  }))
  newGrid.transition = netGrid.transition ?? null
  newGrid.layers = buildLayersFromMessage(netGrid, rowCount, colCount)
  newGrid.rebuildTopTileMap()

  return newGrid
}
