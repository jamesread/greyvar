import { describe, it, expect, vi } from 'vitest'

vi.mock('../menu/menuBorder.js', () => ({}))
const { messageHoldMs } = await import('./FloatingMessageBox.js')

describe('floating message fade timing', () => {
  it('short messages fade after a few seconds, longer ones stay longer, capped', () => {
    expect(messageHoldMs('Got Hay')).toBe(2500 + 7 * 40)
    expect(messageHoldMs('x'.repeat(40))).toBeGreaterThan(messageHoldMs('Got Hay'))
    expect(messageHoldMs('x'.repeat(500))).toBe(6000)
  })
})
