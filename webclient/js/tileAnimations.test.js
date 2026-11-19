import { describe, it, expect } from 'vitest'
import { tileAnimKey, tilesetAnimationsFromMessage, indexTileAnimations, phaserTileAnimConfig, tileAnimStartFrame } from './tileAnimations.js'

describe('tile animations', () => {
  const tileset = {
    key: 'grass',
    animations: tilesetAnimationsFromMessage({
      animations: [
        { tileId: 115, frames: [115, 116, 117], durationsMs: [200, 0, 200] },
        { tileId: 3, frames: [3], durationsMs: [100] }
      ]
    })
  }

  it('normalises messages and drops single-frame animations', () => {
    expect(tileset.animations.map((a) => a.tileId)).toEqual([115])
  })

  it('indexes by atlas and tile id', () => {
    const idx = indexTileAnimations([tileset])
    expect(idx.get(tileAnimKey('grass', 115)).frames).toEqual([115, 116, 117])
    expect(idx.has(tileAnimKey('grass', 3))).toBe(false)
  })

  it('builds a looping Phaser config with per-frame durations', () => {
    const cfg = phaserTileAnimConfig(indexTileAnimations([tileset]).get('tile:grass:115'))
    expect(cfg.repeat).toBe(-1)
    expect(cfg.frames.map((f) => f.duration)).toEqual([200, 100, 200])
    expect(cfg.frames[0]).toMatchObject({ key: 'grass', frame: 115 })
  })

  it('gives stable, varied start frames per cell', () => {
    expect(tileAnimStartFrame(3, 4, 6)).toBe(tileAnimStartFrame(3, 4, 6))
    const starts = new Set()
    for (let r = 0; r < 8; r++) {
      for (let c = 0; c < 8; c++) {
        const s = tileAnimStartFrame(r, c, 4)
        expect(s).toBeGreaterThanOrEqual(0)
        expect(s).toBeLessThan(4)
        starts.add(s)
      }
    }
    expect(starts.size).toBe(4)
    expect(tileAnimStartFrame(1, 1, 1)).toBe(0)
  })
})

describe('Tile.fromNet', () => {
  it('treats a missing frameIndex on an atlas tile as frame 0 (JSON omits zeros)', async () => {
    const { default: Tile } = await import('./Tile.js')
    const from = (net) => { const t = new Tile(); t.fromNet(net); return t }
    expect(from({ atlasKey: 'flora', tex: 'flora.png#0', row: 1, col: 2 }).frameIndex).toBe(0)
    expect(from({ atlasKey: 'flora', frameIndex: 4 }).frameIndex).toBe(4)
    expect(from({ tex: 'water.png' }).frameIndex).toBe(null)
  })
})

describe('synced tile animations', () => {
  it('keeps the synced flag and picks the frame from a shared clock', async () => {
    const { tilesetAnimationsFromMessage, syncedStartFrame } = await import('./tileAnimations.js')
    const [anim] = tilesetAnimationsFromMessage({ animations: [{ tileId: 153, frames: [400, 401, 402, 403], durationsMs: [320, 320, 320, 320], synced: true }] })
    expect(anim.synced).toBe(true)
    expect([0, 319, 320, 700, 1279, 1280].map((t) => syncedStartFrame(anim, t))).toEqual([0, 0, 1, 2, 3, 0])
  })
})

describe('Tile.still', () => {
  it('keeps the server still flag (animated tile drawn static here)', async () => {
    const { default: Tile } = await import('./Tile.js')
    const t = new Tile(); t.fromNet({ atlasKey: 'grass_terains', frameIndex: 44, still: true })
    expect(t.still).toBe(true)
    const u = new Tile(); u.fromNet({ atlasKey: 'grass_terains', frameIndex: 44 })
    expect(u.still).toBe(false)
  })
})
