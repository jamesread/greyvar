import { describe, it, expect } from 'vitest'
import { spriteBounds } from './spriteBounds.js'

describe('spriteBounds', () => {
  it('is the sprite rectangle for plain sprites', () => {
    expect(spriteBounds({ x: 32, y: 48, displayWidth: 16, displayHeight: 16 }))
      .toEqual({ left: 32, top: 48, right: 48, bottom: 64 })
  })

  it('follows a character drawn above and around its cell', () => {
    // adult: 33x33 frame, drawn 8 left and 16 up of the cell at (32, 48)
    expect(spriteBounds({ x: 32, y: 48, displayWidth: 33, displayHeight: 33, offset: { x: -8, y: -16 } }))
      .toEqual({ left: 24, top: 32, right: 57, bottom: 65 })
  })

  it('defaults to one cell', () => {
    expect(spriteBounds({ x: 0, y: 0 })).toEqual({ left: 0, top: 0, right: 16, bottom: 16 })
  })
})
