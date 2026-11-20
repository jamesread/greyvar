// The HUD hotbar's rules, kept free of Phaser so they are easy to test.
// Slots show the inventory in order; the selected slot is the item the local
// player holds (the server is the source of truth: selecting sends an
// appearance request, and the hotbar follows the held item it reports back).

export const HOTBAR_SIZE = 8

/**
 * @param {Array<{definition: string, count?: number}>} inventory
 * @param {Map<string, {title?: string, icon?: string}>} itemDefs
 * @param {string} held item id in hand ('' for none)
 */
export function hotbarSlots (inventory, itemDefs, held, size = HOTBAR_SIZE) {
  const slots = []
  for (let i = 0; i < size; i++) {
    const entry = inventory?.[i]
    if (entry == null || !entry.definition) {
      slots.push({ index: i, item: null, count: 0, title: '', icon: '', selected: false })
      continue
    }
    const def = itemDefs?.get?.(entry.definition)
    slots.push({
      index: i,
      item: entry.definition,
      count: entry.count ?? 1,
      title: def?.title ?? entry.definition,
      icon: def?.icon ?? '',
      selected: held !== '' && held === entry.definition
    })
  }
  return slots
}

/**
 * What pressing a slot asks the server to hold: the slot's item, '-' to put
 * the held item away (pressing the selected slot again), or null for nothing
 * (an empty slot).
 */
export function heldItemForSlot (slots, index) {
  const slot = slots[index]
  if (slot?.item == null) {
    return null
  }
  return slot.selected ? '-' : slot.item
}

/** The item to hold when scrolling the wheel by dir (+1 / -1), wrapping over filled slots. */
export function cycleHeldItem (slots, held, dir) {
  const filled = slots.filter((s) => s.item != null)
  if (filled.length === 0) {
    return null
  }
  const at = filled.findIndex((s) => s.item === held)
  if (at < 0) {
    return (dir > 0 ? filled[0] : filled[filled.length - 1]).item
  }
  return filled[(at + (dir > 0 ? 1 : -1) + filled.length) % filled.length].item
}

/** A string that changes whenever the hotbar needs redrawing. */
export function hotbarSignature (slots) {
  return slots.map((s) => `${s.item ?? ''}:${s.count}:${s.selected ? 1 : 0}:${s.icon}`).join('|')
}
