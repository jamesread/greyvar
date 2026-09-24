import Phaser from 'phaser'
import { initMovementInput } from '../MovementInput.js'
import { createFloatingMessageBox } from '../ui/FloatingMessageBox.js'

export default class HudScene extends Phaser.Scene {
  constructor () {
    super({ key: 'hud' })
  }

  preload () {
    this.load.setBaseURL('/res/')
    this.load.image('menu', 'img/textures/hud/menu.png')
  }

  create () {
    initMovementInput(this)

    const btnMenu = this.add.image(60, 60, 'menu')
    btnMenu.setScrollFactor(0)
    btnMenu.setDepth(100)
    btnMenu.setInteractive({ useHandCursor: true })
    btnMenu.on('pointerdown', this.mnuClicked)

    this.floatingMessage = createFloatingMessageBox(this)

    this.input.keyboard?.on('keydown-ESC', this.onEscape, this)
    this.events.once('shutdown', this.onShutdown, this)

    window.hudScene = this
  }

  update () {
    this.floatingMessage?.update()
  }

  /**
   * Show or clear a floating world message (signs, pickups, etc.).
   * @param {{ text?: string, entityId?: number|string|null, worldX?: number, worldY?: number } | string} msg
   */
  setHudMessage (msg) {
    if (typeof msg === 'string') {
      this.floatingMessage?.show({ text: msg })
      return
    }
    if (msg == null) {
      this.floatingMessage?.hide()
      return
    }
    const text = msg.text ?? ''
    if (String(text).trim() === '') {
      this.floatingMessage?.hide()
      return
    }
    this.floatingMessage?.show({
      text: String(text),
      entityId: msg.entityId != null ? Number(msg.entityId) : null,
      worldX: msg.worldX,
      worldY: msg.worldY
    })
  }

  mnuClicked () {
    window.menuSystem?.toggle('main')
  }

  onEscape () {
    if (window.gameConsole?.isOpen?.()) {
      return
    }
    window.menuSystem?.toggle('main')
  }

  onShutdown () {
    this.input.keyboard?.off('keydown-ESC', this.onEscape, this)
    this.floatingMessage?.destroy()
    this.floatingMessage = null
    if (window.hudScene === this) {
      window.hudScene = null
    }
  }
}
