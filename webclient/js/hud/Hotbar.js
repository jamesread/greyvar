// HUD hotbar: the first HOTBAR_SIZE inventory items along the bottom of the
// screen, with icons and counts. 1-8, the mouse wheel or a click/tap picks
// the item to hold; the selected slot is the held item.
import Phaser from 'phaser'
import { MENU_THEME } from '../menu/MenuTheme.js'
import { localHeldItem } from '../menu/inventoryMenus.js'
import { isTouchDevice, TOUCH_CONTROLS_HEIGHT } from '../TouchControls.js'
import { HOTBAR_SIZE, hotbarSlots, heldItemForSlot, cycleHeldItem, hotbarSignature } from './hotbarModel.js'

const SLOT = 44
const GAP = 6
const MARGIN = 16
const ICON_SCALE = 2
const DEPTH = 90
const LABEL_MS = 1600
const KEY_NAMES = ['ONE', 'TWO', 'THREE', 'FOUR', 'FIVE', 'SIX', 'SEVEN', 'EIGHT']

function inputBlocked () {
  return window.gameConsole?.isOpen?.() ||
    window.menuSystem?.isOpen?.() ||
    !window.phaser?.scene?.isActive('grid')
}

export default class Hotbar {
  constructor (scene) {
    this.scene = scene
    this.slots = []
    this.signature = null
    this.labelTimer = null
    this.root = scene.add.container(0, 0).setScrollFactor(0).setDepth(DEPTH)
    this.slotViews = []
    for (let i = 0; i < HOTBAR_SIZE; i++) {
      this.slotViews.push(this.makeSlot(i))
    }
    this.label = scene.add.text(0, 0, '', {
      fontFamily: MENU_THEME.fontFamily,
      fontSize: '14px',
      color: MENU_THEME.title,
      backgroundColor: '#252b4acc',
      padding: { x: 8, y: 3 }
    }).setOrigin(0.5, 1).setScrollFactor(0).setDepth(DEPTH).setVisible(false)

    this.keys = KEY_NAMES.map((name, i) => {
      const key = scene.input.keyboard?.addKey(Phaser.Input.Keyboard.KeyCodes[name])
      key?.on('down', () => this.press(i))
      return key
    })
    this.onWheel = (_pointer, _over, _dx, dy) => {
      if (dy !== 0 && !inputBlocked()) {
        this.hold(cycleHeldItem(this.slots, localHeldItem(), dy > 0 ? 1 : -1))
      }
    }
    scene.input.on('wheel', this.onWheel)
    this.onResize = () => this.layout()
    scene.scale.on('resize', this.onResize)
    this.layout()
    this.refresh()
  }

  makeSlot (i) {
    const scene = this.scene
    const bg = scene.add.rectangle(0, 0, SLOT, SLOT, MENU_THEME.panel, 0.82)
      .setStrokeStyle(2, MENU_THEME.panelBorderDark, 0.9)
      .setInteractive({ useHandCursor: true })
    bg.on('pointerdown', () => this.press(i))
    const icon = scene.add.image(0, 0, '__DEFAULT').setScale(ICON_SCALE).setVisible(false)
    const count = scene.add.text(SLOT / 2 - 4, SLOT / 2 - 2, '', {
      fontFamily: MENU_THEME.fontFamily, fontSize: '12px', color: '#ffffff', stroke: '#000000', strokeThickness: 3
    }).setOrigin(1, 1)
    const number = scene.add.text(-SLOT / 2 + 4, -SLOT / 2 + 2, String(i + 1), {
      fontFamily: MENU_THEME.fontFamily, fontSize: '10px', color: MENU_THEME.subtitle
    }).setOrigin(0, 0)
    const view = scene.add.container(0, 0, [bg, icon, count, number])
    this.root.add(view)
    return { view, bg, icon, count }
  }

  layout () {
    const { width, height } = this.scene.scale
    // Shrink to fit narrow (phone) screens; with touch controls, sit just
    // above the D-pad and buttons instead of on top of them.
    const full = HOTBAR_SIZE * SLOT + (HOTBAR_SIZE - 1) * GAP
    const scale = Math.min(1, (width - 2 * MARGIN) / full)
    const slot = SLOT * scale
    const gap = GAP * scale
    const total = HOTBAR_SIZE * slot + (HOTBAR_SIZE - 1) * gap
    const bottom = isTouchDevice() ? TOUCH_CONTROLS_HEIGHT + 20 : MARGIN
    const y = height - bottom - slot / 2
    this.slotViews.forEach((v, i) => {
      v.view.setScale(scale).setPosition((width - total) / 2 + slot / 2 + i * (slot + gap), y)
    })
    this.label.setPosition(width / 2, y - slot / 2 - 8)
  }

  /** Re-read inventory and held item; redraw only when something changed. */
  refresh () {
    const gs = window.gameState
    const held = localHeldItem()
    const slots = hotbarSlots(gs?.inventory ?? [], gs?.itemDefinitions, held)
    const sig = hotbarSignature(slots)
    if (sig === this.signature) {
      return
    }
    const before = this.slots.find((s) => s.selected)?.item ?? null
    this.signature = sig
    this.slots = slots
    slots.forEach((slot, i) => this.drawSlot(this.slotViews[i], slot))
    const now = slots.find((s) => s.selected)
    if ((now?.item ?? null) !== before && this.signature != null) {
      this.showLabel(now ? now.title : before ? 'Hands empty' : '')
    }
  }

  drawSlot (v, slot) {
    v.bg.setStrokeStyle(slot.selected ? 3 : 2, slot.selected ? MENU_THEME.panelBorder : MENU_THEME.panelBorderDark, slot.selected ? 1 : 0.9)
    v.bg.setFillStyle(slot.selected ? MENU_THEME.focusFill : MENU_THEME.panel, slot.selected ? 0.95 : 0.82)
    v.count.setText(slot.item != null && slot.count > 1 ? String(slot.count) : '')
    if (slot.item == null || !slot.icon) {
      v.icon.setVisible(false)
      return
    }
    const key = slot.icon
    if (this.scene.textures.exists(key)) {
      v.icon.setTexture(key).setVisible(true)
      v.icon.texture.setFilter(Phaser.Textures.FilterMode.NEAREST)
      return
    }
    v.icon.setVisible(false)
    this.scene.load.image(key, key)
    this.scene.load.once(`filecomplete-image-${key}`, () => {
      this.signature = null // redraw with the icon now loaded
    })
    this.scene.load.start()
  }

  showLabel (text) {
    if (!text) {
      return
    }
    this.label.setText(text).setVisible(true)
    this.labelTimer?.remove(false)
    this.labelTimer = this.scene.time.delayedCall(LABEL_MS, () => this.label.setVisible(false))
  }

  press (index) {
    if (inputBlocked()) {
      return
    }
    this.hold(heldItemForSlot(this.slots, index))
  }

  hold (heldItem) {
    if (heldItem == null) {
      return
    }
    window.serverConnection?.sendAppearanceRequest({ heldItem })
  }

  destroy () {
    this.scene.input.off('wheel', this.onWheel)
    this.scene.scale.off('resize', this.onResize)
    for (const key of this.keys) {
      key?.removeAllListeners()
    }
    this.labelTimer?.remove(false)
    this.root.destroy(true)
    this.label.destroy()
  }
}
