// When the local player is behind something drawn over characters (tiles
// marked above, e.g. a tree's canopy), show where they are: a shaded copy of
// the hidden pixels with a light outline, drawn over the canopy and masked to
// it, so it appears only where the player is actually covered.
import Phaser from 'phaser'

const OUTLINE_COLOR = 0xe8f4ff
const OUTLINE_ALPHA = 0.85
const SHADE_TINT = 0x6f7fb0
const SHADE_ALPHA = 0.6
const OFFSETS = [[-1, 0], [1, 0], [0, -1], [0, 1]]

export default class OcclusionGhost {
  /**
   * @param {Phaser.Scene} scene
   * @param {Phaser.GameObjects.Layer} occluders what can hide the player
   */
  constructor (scene, occluders) {
    this.scene = scene
    this.enabled = scene.sys.game.renderer?.type === Phaser.WEBGL
    this.root = scene.add.container(0, 0).setVisible(false)
    this.outline = scene.add.container(0, 0)
    this.shade = scene.add.container(0, 0)
    this.root.add([this.outline, this.shade])
    if (this.enabled) {
      this.root.enableFilters()
      this.root.filters.external.addMask(occluders)
    }
  }

  /**
   * Follow a LayeredSprite; show the ghost only while covered is true.
   * @param {Phaser.GameObjects.Container|null} sprite
   * @param {boolean} covered whether the sprite overlaps an occluding cell
   */
  update (sprite, covered) {
    const show = this.enabled && covered && sprite?.active === true && sprite.visible
    this.root.setVisible(show)
    if (!show) {
      return
    }
    this.root.setPosition(sprite.x, sprite.y)
    const parts = sprite.list.filter((o) => o.texture != null)
    this.sync(this.shade, parts, 1, (img) => {
      img.setTintMode(Phaser.TintModes.MULTIPLY).setTint(SHADE_TINT).setAlpha(SHADE_ALPHA)
    })
    this.sync(this.outline, parts, OFFSETS.length, (img) => {
      img.setTintMode(Phaser.TintModes.FILL).setTint(OUTLINE_COLOR).setAlpha(OUTLINE_ALPHA)
    })
  }

  // Mirror parts into `holder` (copies per part, offset for the outline).
  sync (holder, parts, copies, style) {
    const need = parts.length * copies
    while (holder.list.length < need) {
      const img = this.scene.add.image(0, 0, '__DEFAULT').setOrigin(0)
      style(img)
      holder.add(img)
    }
    holder.list.forEach((img, i) => {
      const part = parts[Math.floor(i / copies)]
      if (part == null) {
        img.setVisible(false)
        return
      }
      const [dx, dy] = copies > 1 ? OFFSETS[i % copies] : [0, 0]
      if (img.texture.key !== part.texture.key) {
        img.setTexture(part.texture.key)
      }
      if (img.frame.name !== part.frame.name) {
        img.setFrame(part.frame.name)
      }
      img.setPosition(part.x + dx, part.y + dy)
      img.setFlip(part.flipX, part.flipY)
      img.setVisible(part.visible)
    })
  }

  destroy () {
    this.root.destroy(true)
  }
}
