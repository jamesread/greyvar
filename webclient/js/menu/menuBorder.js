import Phaser from 'phaser'

export const MENU_BORDER_KEY = 'menu-border'
export const MENU_BORDER_PATH = 'img/textures/hud/menu-border.png'
export const MENU_BORDER_SLICE = 16

export function preloadMenuBorder (loader) {
  loader.image(MENU_BORDER_KEY, MENU_BORDER_PATH)
}

export function menuBorderReady (scene) {
  return scene?.textures?.exists(MENU_BORDER_KEY) ?? false
}

export function configureMenuBorderTexture (scene) {
  if (!menuBorderReady(scene)) {
    return
  }

  const texture = scene.textures.get(MENU_BORDER_KEY)
  texture.setFilter(Phaser.Textures.FilterMode.NEAREST)
}

export function createMenuBorderPanel (scene, x, y, width, height) {
  if (!menuBorderReady(scene)) {
    return null
  }

  configureMenuBorderTexture(scene)
  const slice = MENU_BORDER_SLICE
  const panel = scene.add.nineslice(
    x,
    y,
    MENU_BORDER_KEY,
    undefined,
    width,
    height,
    slice,
    slice,
    slice,
    slice
  )
  panel.setOrigin(0, 0)
  return panel
}

export function layoutMenuBorderPanel (panel, x, y, width, height) {
  if (panel == null) {
    return
  }

  panel.setPosition(x, y)
  panel.setSize(width, height)
}

export function drawMenuBorderFallback (graphics, x, y, width, height, theme) {
  graphics.clear()
  graphics.fillStyle(theme.panel, 1)
  graphics.fillRoundedRect(x, y, width, height, 8)
  graphics.lineStyle(2, theme.panelBorderDark, 1)
  graphics.strokeRoundedRect(x, y, width, height, 8)
  graphics.lineStyle(1, theme.panelBorder, 0.9)
  graphics.strokeRoundedRect(x + 3, y + 3, width - 6, height - 6, 6)
}
