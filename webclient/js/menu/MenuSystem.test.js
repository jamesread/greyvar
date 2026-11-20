import { describe, it, expect, vi } from 'vitest'

globalThis.window = globalThis.window ?? {}
vi.mock('phaser', () => ({ default: {} }))
vi.mock('./menuBorder.js', () => ({}))
vi.mock('./inventoryMenus.js', () => ({ registerInventoryMenus: () => {}, openInventoryMenu: () => {}, openCraftingMenu: () => {} }))
const { default: MenuSystem } = await import('./MenuSystem.js')

function openMenus (...ids) {
  const ms = new MenuSystem({ scene: { isActive: () => ms.stack.length > 0 } })
  ms.render = () => {}
  ms.ensureSceneRunning = () => {}
  ms.ui = { itemRows: [] }
  for (const id of ids) {
    ms.registerMenu({ id, items: [] })
  }
  ms.open(ids[0])
  for (const id of ids.slice(1)) {
    ms.navigateTo(id)
  }
  ms.close = () => { ms.stack = [] }
  return ms
}
const press = (ms, code) => {
  const event = { code, key: code, preventDefault: () => {} }
  ms.onMenuKeyDown(event)
  return event
}

describe('menu keys', () => {
  for (const code of ['Escape', 'Backspace']) {
    it(`${code} goes back a level, then closes`, () => {
      const ms = openMenus('main', 'inventory')
      const event = press(ms, code)
      expect(ms.stack).toEqual(['main'])
      expect(event.greyvarMenuHandled).toBe(true)
      press(ms, code)
      expect(ms.stack).toEqual([])
    })
  }

  it('Left only goes back inside a submenu (never closes)', () => {
    const ms = openMenus('main', 'inventory')
    press(ms, 'ArrowLeft')
    expect(ms.stack).toEqual(['main'])
    press(ms, 'ArrowLeft')
    expect(ms.stack).toEqual(['main'])
  })
})
