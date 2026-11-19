import { describe, it, expect } from 'vitest'
import { formatDebugPos } from './debugPos.js'

describe('formatDebugPos', () => {
  it('shows pixels, the tile under the player and the map', () => {
    expect(formatDebugPos({ x: 112, y: 128 }, { worldId: 'isleOfStarting_dev_tiled', gridId: '1.1.tmj' }))
      .toBe('x 112  y 128  |  tile row 8 col 7  |  isleOfStarting_dev_tiled / 1.1.tmj')
  })
  it('rounds mid-step sprite positions and copes with no player or map', () => {
    expect(formatDebugPos({ x: 20.6, y: 7.2 }, null)).toBe('x 21  y 7  |  tile row 0 col 1')
    expect(formatDebugPos(null, null)).toBe('pos: no player')
  })
})
