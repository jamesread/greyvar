import { describe, it, expect, vi } from 'vitest'

vi.mock('phaser', () => ({ default: {} }))
const { pickInteractTarget } = await import('./InteractInput.js')

describe('pickInteractTarget', () => {
  const me = 1
  const entities = {
    1: { entityId: 1, x: 112, y: 70 },
    2: { entityId: 2, x: 112, y: 48 }, // Bill, straight ahead (up), 22px
    3: { entityId: 3, x: 128, y: 88 } // a sign beside and behind, 18px
  }

  it('prefers what the player faces over what is merely nearest', () => {
    expect(pickInteractTarget(entities, me, 112, 70, 'Up').entityId).toBe(2)
    expect(pickInteractTarget(entities, me, 112, 70, 'Down').entityId).toBe(3)
  })

  it('only offers entities within the server range (24px)', () => {
    expect(pickInteractTarget(entities, me, 112, 75, 'Up').entityId).toBe(3)
    expect(pickInteractTarget({ 1: entities[1] }, me, 112, 70, 'Up')).toBe(null)
  })

  it('reaches further for entities with an interactRange (across a counter)', () => {
    const counter = { 1: { entityId: 1, x: 112, y: 80 }, 4: { entityId: 4, definition: 'herman', x: 112, y: 48 } }
    expect(pickInteractTarget(counter, me, 112, 80, 'Up')).toBe(null)
    const rangeFor = (ent) => (ent.definition === 'herman' ? 40 : 0)
    expect(pickInteractTarget(counter, me, 112, 80, 'Up', rangeFor).entityId).toBe(4)
  })
})
