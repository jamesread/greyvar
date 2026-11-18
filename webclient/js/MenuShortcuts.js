import { openCraftingMenu, openInventoryMenu } from './menu/inventoryMenus.js'
import { isTouchDevice } from './TouchControls.js'

let sceneRef = null

function shortcutsBlocked () {
  return (
    window.gameConsole?.isOpen?.() ||
    !window.phaser?.scene?.isActive('grid')
  )
}

export function initMenuShortcuts (scene) {
  if (isTouchDevice() || scene?.input?.keyboard == null) {
    return
  }

  sceneRef = scene
  const keyboard = scene.input.keyboard
  keyboard.addCapture(['TAB', 'C'])
  keyboard.on('keydown-TAB', onInventoryKey, scene)
  keyboard.on('keydown-C', onCraftingKey, scene)
  scene.events.once('shutdown', destroyMenuShortcuts)
}

function onInventoryKey (event) {
  if (shortcutsBlocked() || window.menuSystem?.isOpen()) {
    return
  }

  event.preventDefault()
  openInventoryMenu()
}

function onCraftingKey () {
  if (shortcutsBlocked() || window.menuSystem?.isOpen()) {
    return
  }

  openCraftingMenu()
}

export function destroyMenuShortcuts () {
  if (sceneRef?.input?.keyboard != null) {
    sceneRef.input.keyboard.off('keydown-TAB', onInventoryKey, sceneRef)
    sceneRef.input.keyboard.off('keydown-C', onCraftingKey, sceneRef)
  }

  sceneRef = null
}
