import { MENU_THEME } from './menu/MenuTheme.js'
import { openCraftingMenu, openInventoryMenu } from './menu/inventoryMenus.js'
import { isTouchDevice } from './TouchControls.js'

const BUTTON_W = 52
const BUTTON_H = 44
const BUTTON_GAP = 8
const MARGIN = 20
const DEPTH = 100

let container = null
let sceneRef = null

function makeLabelButton (scene, label, onPress) {
  const btn = scene.add.container(0, 0)

  const bg = scene.add.rectangle(0, 0, BUTTON_W, BUTTON_H, MENU_THEME.panel, 0.85)
  bg.setStrokeStyle(2, MENU_THEME.panelBorder, 0.9)
  bg.setInteractive({ useHandCursor: true })

  const text = scene.add.text(0, 0, label, {
    fontFamily: MENU_THEME.fontFamily,
    fontSize: '13px',
    color: MENU_THEME.item
  }).setOrigin(0.5)

  bg.on('pointerdown', () => {
    bg.setAlpha(1)
    onPress()
  })
  bg.on('pointerup', () => bg.setAlpha(0.85))
  bg.on('pointerout', () => bg.setAlpha(0.85))

  btn.add([bg, text])
  btn.setScrollFactor(0)
  btn.setDepth(DEPTH)
  return btn
}

function layoutButtons () {
  if (container == null || sceneRef == null) {
    return
  }

  const x = sceneRef.scale.width - MARGIN - BUTTON_W / 2
  const invY = sceneRef.scale.height - MARGIN - BUTTON_H / 2
  const craftY = invY - BUTTON_H - BUTTON_GAP

  const children = container.list
  if (children.length >= 2) {
    children[0].setPosition(x, craftY)
    children[1].setPosition(x, invY)
  }
}

export function initHudShortcutButtons (scene) {
  if (!isTouchDevice() || container != null) {
    return
  }

  sceneRef = scene
  container = scene.add.container(0, 0)
  container.setScrollFactor(0)
  container.setDepth(DEPTH)

  const craftBtn = makeLabelButton(scene, 'Craft', openCraftingMenu)
  const invBtn = makeLabelButton(scene, 'Items', openInventoryMenu)
  container.add([craftBtn, invBtn])

  scene.scale.on('resize', layoutButtons)
  scene.events.once('shutdown', destroyHudShortcutButtons)

  layoutButtons()
}

export function destroyHudShortcutButtons () {
  if (sceneRef != null) {
    sceneRef.scale?.off('resize', layoutButtons)
  }

  container?.destroy(true)
  container = null
  sceneRef = null
}
