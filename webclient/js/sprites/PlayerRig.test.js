import { describe, it, expect, vi } from 'vitest'

vi.mock('phaser', () => ({ default: {} }))
const { sizeOf, rigKey, DEFAULT_SIZE, RIG_URLS } = await import('./PlayerRig.js')

describe('character sizes', () => {
  it('draws adults unless a look says child', () => {
    expect(DEFAULT_SIZE).toBe('adult')
    expect(sizeOf(null)).toBe('adult')
    expect(sizeOf({ rig: '' })).toBe('adult')
    expect(sizeOf({ rig: 'child' })).toBe('child')
    expect(rigKey('child')).toBe('rig:player')
    expect(rigKey('adult')).toBe('rig:adult')
    expect(Object.keys(RIG_URLS).sort()).toEqual(['adult', 'child'])
  })
})
