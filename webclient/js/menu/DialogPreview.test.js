import { describe, it, expect, vi } from 'vitest'

vi.mock('phaser', () => ({ default: { GameObjects: { Container: class {} } } }))
const { previewScale, previewSpritePosition } = await import('./DialogPreview.js')

describe('dialog preview', () => {
  it('keeps the child at 8x and fits the adult frame in about the same space', () => {
    expect(previewScale(16)).toBe(8)
    expect(previewScale(33)).toBe(4)
    expect(33 * previewScale(33)).toBeLessThanOrEqual(16 * previewScale(16) + 8)
  })

  it('cancels the sprite anchor offset so the frame sits in the panel', () => {
    // adult: drawn 8 px left and 16 px up of its cell
    expect(previewSpritePosition(100, 50, { x: -8, y: -16 }, 4)).toEqual({ x: 132, y: 114 })
    // child: no offset
    expect(previewSpritePosition(100, 50, { x: 0, y: 0 }, 8)).toEqual({ x: 100, y: 50 })
    expect(previewSpritePosition(100, 50, undefined, 8)).toEqual({ x: 100, y: 50 })
  })
})
