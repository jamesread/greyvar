// A large live preview of the local player beside a customization dialog
// (tailor, skin tone, hairdresser): it shows the player as they would look
// with the focused option applied.
import { MENU_THEME } from './MenuTheme.js'
import LayeredSprite from '../sprites/LayeredSprite.js'
import { getRig, sizeOf } from '../sprites/PlayerRig.js'
import { applyWear } from '../sprites/appearance.js'

const PAD = 14
const GAP = 16

// The preview's size on screen (about 130 px whatever the character size):
// 8x for a 16 px child, 4x for a 33 px adult.
const PREVIEW_PX = 136

/** Whole-number scale for a rig's frame size. */
export function previewScale (frameSize) {
  return Math.max(2, Math.floor(PREVIEW_PX / frameSize))
}

/**
 * Where to put the sprite so its frame's top-left lands at (x, y): a
 * LayeredSprite draws its layers at its anchor offset (an adult is drawn
 * up and left of its cell), and that offset is scaled with it.
 */
export function previewSpritePosition (x, y, offset, scale) {
  return { x: x - (offset?.x ?? 0) * scale, y: y - (offset?.y ?? 0) * scale }
}

/** The local player's current look, from their sprite. */
export function localAppearance () {
  const gs = window.gameState
  const ent = gs?.gridScene?.entities?.[gs?.localPlayerEntityId]
  return ent?.img?.appearance ?? ent?.appearance ?? null
}

/**
 * @param {{scene: Phaser.Scene, root: Phaser.GameObjects.Container}} ctx
 * @param {(itemId: string) => object|null} wearFor the wear of a menu item (null: none)
 */
export function createDialogPreview ({ scene, root }, wearFor) {
  const base = localAppearance()
  const rig = getRig(scene, sizeOf(base))
  if (rig == null || base == null) {
    return null
  }
  const scale = previewScale(rig.frameSize)
  const size = rig.frameSize * scale
  const frame = scene.add.rectangle(0, 0, size + PAD * 2, size + PAD * 2 + 22, MENU_THEME.panel, 0.95)
    .setOrigin(0).setStrokeStyle(2, MENU_THEME.panelBorder, 1)
  const label = scene.add.text(0, 0, 'Preview', {
    fontFamily: MENU_THEME.fontFamily, fontSize: '12px', color: MENU_THEME.subtitle
  }).setOrigin(0.5, 0)
  const sprite = new LayeredSprite(scene, 0, 0, rig, base, rig.initialState,
    (itemId) => window.gameState?.itemDefinitions?.get(itemId))
  sprite.setScale(scale)
  root.add([frame, label, sprite])

  return {
    layout ({ panelX, panelY, panelW, width }) {
      const w = size + PAD * 2
      let x = panelX - GAP - w
      if (x < 8) {
        x = panelX + panelW + GAP // no room on the left: try the right
      }
      const fits = x >= 8 && x + w <= width - 8
      for (const o of [frame, label, sprite]) {
        o.setVisible(fits)
      }
      frame.setPosition(x, panelY)
      label.setPosition(x + w / 2, panelY + 6)
      const at = previewSpritePosition(x + PAD, panelY + 22 + PAD, sprite.offset, scale)
      sprite.setPosition(at.x, at.y)
    },
    focus (itemId) {
      sprite.setAppearance(applyWear(rig, base, wearFor(itemId)))
    }
  }
}
