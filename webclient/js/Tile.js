export default class Tile {
  constructor(netTile) {
    this.row = -1
    this.col = -1

    this.textureName = 'construct.png'
    this.textureHorizontalFlip = false
    this.textureVerticalFlip = false
    this.textureRotation = 0
    this.atlasKey = null
    this.frameIndex = null
    this.collision = []
    this.collisionFromTsx = false
  }

  fromNet (netTile) {
    this.row = netTile.row ?? 0
    this.col = netTile.col ?? 0

    this.textureName = netTile.tex
    this.textureHorizontalFlip = netTile.flipH ?? false
    this.textureVerticalFlip = netTile.flipV ?? false
    this.textureRotation = netTile.rot ?? 0
    this.atlasKey = netTile.atlasKey ?? null
    this.frameIndex = netTile.frameIndex ?? null
    this.collisionFromTsx = netTile.collisionFromTsx ?? false
    this.collision = (netTile.collision ?? []).map((rect) => ({
      x: rect.x ?? 0,
      y: rect.y ?? 0,
      w: rect.w ?? 0,
      h: rect.h ?? 0
    }))
  }
}
