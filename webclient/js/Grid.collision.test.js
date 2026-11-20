import { describe, it, expect } from 'vitest'
import Grid from './Grid.js'

const tile = (row, col, extra = {}) => ({ row, col, collision: [], ...extra })

function gridWith (...layers) {
  const g = new Grid()
  g.layers = layers.map((tiles) => ({ tiles }))
  return g
}

describe('Grid.collisionTileAt (mirrors the server)', () => {
  const rim = tile(0, 0, { textureName: 'rim', collision: [{ x: 6, y: 0, w: 4, h: 16 }] })
  const cap = tile(0, 0, { overlay: true })
  const flower = tile(0, 1, { decor: true })
  const sand = tile(0, 1, { textureName: 'sand' })
  const ext = tile(0, 1, { overlay: true, collision: [{ x: 0, y: 0, w: 16, h: 12 }] })

  it('sees the hill tile under a generated face cap', () => {
    const g = gridWith([rim, sand], [cap, ext])
    expect(g.topTileAt(0, 0)).toBe(cap)
    expect(g.collisionTileAt(0, 0)).toBe(rim)
  })

  it('adds an extension\'s rects to the tile beneath, without changing it', () => {
    const got = gridWith([rim, sand], [cap, ext]).collisionTileAt(0, 1)
    expect(got.textureName).toBe('sand')
    expect(got.collision).toEqual([{ x: 0, y: 0, w: 16, h: 12 }])
    expect(sand.collision).toEqual([])
  })

  it('looks through decor', () => {
    expect(gridWith([sand], [flower]).collisionTileAt(0, 1)).toBe(sand)
  })

  it('adds an overlay drawn under a nearer wall (a tall face\'s lower half)', () => {
    const lower = tile(0, 1, { overlay: true, collision: [{ x: 0, y: 0, w: 16, h: 6 }] })
    const wall = tile(0, 1, { textureName: 'wall', collision: [{ x: 0, y: 6, w: 16, h: 4 }] })
    const got = gridWith([sand], [lower], [wall]).collisionTileAt(0, 1)
    expect(got.textureName).toBe('wall')
    expect(got.collision).toEqual([{ x: 0, y: 6, w: 16, h: 4 }, { x: 0, y: 0, w: 16, h: 6 }])
  })
})
