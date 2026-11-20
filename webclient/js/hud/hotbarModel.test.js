import { describe, it, expect } from 'vitest'
import { hotbarSlots, heldItemForSlot, cycleHeldItem, hotbarSignature, HOTBAR_SIZE } from './hotbarModel.js'

const defs = new Map([
  ['hay_seeds', { title: 'Hay Seeds', icon: 'img/textures/items/icons/hay_seeds.png' }],
  ['key', { title: 'Key', icon: 'img/textures/items/icons/key.png' }]
])
const inventory = [{ definition: 'hay_seeds', count: 10 }, { definition: 'key', count: 1 }]

describe('hotbar', () => {
  it('fills slots from the inventory and marks the held item', () => {
    const slots = hotbarSlots(inventory, defs, 'key')
    expect(slots).toHaveLength(HOTBAR_SIZE)
    expect(slots[0]).toMatchObject({ item: 'hay_seeds', count: 10, title: 'Hay Seeds', selected: false })
    expect(slots[1]).toMatchObject({ item: 'key', selected: true, icon: 'img/textures/items/icons/key.png' })
    expect(slots[2].item).toBe(null)
  })

  it('pressing a slot holds it, pressing the held slot puts it away, empty does nothing', () => {
    const slots = hotbarSlots(inventory, defs, 'key')
    expect(heldItemForSlot(slots, 0)).toBe('hay_seeds')
    expect(heldItemForSlot(slots, 1)).toBe('-')
    expect(heldItemForSlot(slots, 5)).toBe(null)
  })

  it('the wheel cycles through filled slots and wraps', () => {
    const slots = hotbarSlots(inventory, defs, '')
    expect(cycleHeldItem(slots, '', 1)).toBe('hay_seeds')
    expect(cycleHeldItem(slots, '', -1)).toBe('key')
    expect(cycleHeldItem(slots, 'hay_seeds', 1)).toBe('key')
    expect(cycleHeldItem(slots, 'key', 1)).toBe('hay_seeds')
    expect(cycleHeldItem(hotbarSlots([], defs, ''), '', 1)).toBe(null)
  })

  it('redraws only when something visible changes', () => {
    const a = hotbarSignature(hotbarSlots(inventory, defs, ''))
    expect(hotbarSignature(hotbarSlots(inventory, defs, ''))).toBe(a)
    expect(hotbarSignature(hotbarSlots(inventory, defs, 'key'))).not.toBe(a)
    expect(hotbarSignature(hotbarSlots([{ definition: 'hay_seeds', count: 9 }, inventory[1]], defs, ''))).not.toBe(a)
  })
})
