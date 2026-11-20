// A character drawn as stacked layer sprites (body, shirt, hair, held item...)
// that all show the same pose, driven by one AnimationController.
import Phaser from 'phaser'
import AnimationController from './AnimationController.js'
import IdleScheduler from './IdleScheduler.js'
import { resolveAppearance } from './appearance.js'
import { layerTexture } from './PlayerRig.js'
import { facingFromState, idleClipForFacing, useClipForFacing } from './facing.js'

// Walking and bumping are transient: remote players only get move/bump
// messages, never an explicit "stopped", so fall back to idle when none has
// arrived for this long. Bumps arrive every ~67ms while a player pushes.
const WALK_IDLE_TIMEOUT_MS = 250

// The world grid cell a character stands in.
const CELL = 16

/** States driven by a stream of events rather than set once. */
export function isTransientState (state) {
  const s = String(state)
  return s.startsWith('walk') || s.startsWith('bump')
}

export default class LayeredSprite extends Phaser.GameObjects.Container {
  /**
   * @param {Phaser.Scene} scene
   * @param {object} rig parsed rig.json
   * @param {object} [appearance] {parts, palettes, heldItem}
   * @param {string} [state] initial entity state
   * @param {(itemId: string) => ({heldSprite?: string, heldGrip?: number[]}|undefined)} [itemResolver]
   */
  constructor (scene, x, y, rig, appearance, state, itemResolver = () => undefined) {
    super(scene, x, y)
    this.rig = rig
    this.itemResolver = itemResolver
    this.heldItemId = ''
    this.setSize(rig.frameSize, rig.frameSize)

    // The frame's anchor (feet) sits on the bottom centre of the 16px cell
    // the container is placed at, so taller rigs rise above it.
    const anchor = rig.anchor ?? [rig.frameSize / 2, rig.frameSize]
    this.offset = { x: CELL / 2 - anchor[0], y: CELL - anchor[1] }

    this.layerImages = new Map()
    for (const layer of rig.layers) {
      const img = new Phaser.GameObjects.Image(scene, this.offset.x, this.offset.y, layer.sheets[layer.default], 0).setOrigin(0)
      this.layerImages.set(layer.name, img)
      this.add(img)
    }
    this.itemImage = null

    this.poseByFrame = new Map(rig.poses.map((p) => [p.frame, p]))
    this.controller = new AnimationController(rig.clips)
    this.idle = new IdleScheduler(rig.idle, this.controller)
    this.state = null
    this.facing = null
    this.sinceWalk = 0

    this.setAppearance(appearance)
    this.playState(state ?? rig.initialState)
    scene.sys.updateList.add(this)
  }

  setAppearance (appearance) {
    this.appearance = resolveAppearance(this.rig, appearance)
    for (const layer of this.rig.layers) {
      const key = layerTexture(this.scene, this.rig, layer, this.appearance.parts[layer.name], this.appearance.palettes)
      this.layerImages.get(layer.name).setTexture(key, this.controller.frame)
    }
    this.syncHeldItem()
    this.applyFrame()
  }

  /** Show appearance.heldItem using its itemdef's held sprite (loaded on demand). */
  syncHeldItem () {
    const id = this.appearance.heldItem ?? ''
    if (id === this.heldItemId) {
      return
    }
    this.heldItemId = id
    const def = id !== '' ? this.itemResolver(id) : undefined
    const sprite = def?.heldSprite
    if (!sprite) {
      this.setHeldItem(null)
      return
    }
    const item = { texture: sprite, grip: def.heldGrip ?? [0, 0] }
    if (this.scene.textures.exists(sprite)) {
      this.setHeldItem(item)
      return
    }
    this.scene.load.once(`filecomplete-image-${sprite}`, () => {
      if (this.active && this.heldItemId === id) {
        this.setHeldItem(item)
      }
    })
    this.scene.load.image(sprite, sprite)
    this.scene.load.start()
  }

  /** Optional held item: {texture, frame?, grip: [x, y]} or null. */
  setHeldItem (item) {
    if (this.itemImage != null) {
      this.itemImage.destroy()
      this.itemImage = null
    }
    this.heldItem = item
    if (item == null || !this.scene.textures.exists(item.texture)) {
      return
    }
    this.itemImage = new Phaser.GameObjects.Image(this.scene, 0, 0, item.texture, item.frame).setOrigin(0)
    this.add(this.itemImage)
    this.applyFrame()
  }

  /** Map an entity state (idle, walkLeft...) to its clip and play it. */
  playState (state) {
    if (isTransientState(state)) {
      this.sinceWalk = 0
    }
    if (state === this.state) {
      return
    }
    // A use gesture finishes by itself and then idles; only movement
    // (walk/bump) cuts it short. The scene asks for idle every frame.
    if (this.state === 'use' && state === this.rig.initialState) {
      return
    }
    this.state = state
    this.facing = facingFromState(state) ?? this.facing
    if (state === this.rig.initialState) {
      // Stand facing the way we last walked; only the front idle has the
      // random variants (blink, glance, shrug).
      const facingClip = idleClipForFacing(this.rig.clips, this.facing)
      if (facingClip != null) {
        this.idle.stop()
        this.controller.play(facingClip, { restart: true })
      } else if (this.rig.idle?.base) {
        this.idle.start()
      }
    } else {
      this.idle.stop()
      this.controller.play(this.rig.states[state] ?? state)
    }
    this.applyFrame()
  }

  /** Turn to face a direction and stand idle (e.g. an NPC being talked to). */
  faceTo (facing) {
    this.facing = facing ?? this.facing
    this.state = null
    this.playState(this.rig.initialState)
  }

  /**
   * Play the one-shot use (action key) animation for a facing, then return
   * to the facing idle. A move arriving meanwhile takes over as usual.
   */
  playUse (facing) {
    const clip = useClipForFacing(this.rig.clips, facing ?? this.facing)
    if (clip == null) {
      return false
    }
    this.facing = facing ?? this.facing
    this.idle.stop()
    this.state = 'use'
    this.controller.play(clip, {
      loop: false,
      restart: true,
      onComplete: () => {
        if (this.state === 'use') {
          this.state = null
          this.playState(this.rig.initialState)
        }
      }
    })
    this.applyFrame()
    return true
  }

  applyFrame () {
    const frame = this.controller.frame
    for (const img of this.layerImages.values()) {
      img.setFrame(frame)
    }
    if (this.itemImage != null) {
      const pose = this.poseByFrame.get(frame)
      const hand = pose?.hand
      this.itemImage.setVisible(hand != null)
      if (hand != null) {
        const grip = this.heldItem?.grip ?? [0, 0]
        this.itemImage.setPosition(this.offset.x + hand[0] - grip[0], this.offset.y + hand[1] - grip[1])
        if (pose.itemZ === 'behind') {
          this.sendToBack(this.itemImage)
        } else {
          this.bringToTop(this.itemImage)
        }
      }
    }
  }

  preUpdate (_time, delta) {
    if (isTransientState(this.state)) {
      this.sinceWalk += delta
      if (this.sinceWalk > WALK_IDLE_TIMEOUT_MS) {
        this.playState(this.rig.initialState)
      }
    }
    this.idle.tick(delta)
    if (this.controller.tick(delta)) {
      this.applyFrame()
    }
  }

  destroy (fromScene) {
    this.scene?.sys.updateList.remove(this)
    super.destroy(fromScene)
  }
}
