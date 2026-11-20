import Phaser from 'phaser'
import { initMovementInput } from '../MovementInput.js'
import { initInteractInput } from '../InteractInput.js'
import { initHudShortcutButtons } from '../HudShortcutButtons.js'
import { initMenuShortcuts } from '../MenuShortcuts.js'
import { initTouchControls } from '../TouchControls.js'
import { createFloatingMessageBox } from '../ui/FloatingMessageBox.js'
import { preloadMenuBorder } from '../menu/menuBorder.js'
import Hotbar from '../hud/Hotbar.js'

export default class HudScene extends Phaser.Scene {
  constructor () {
    super({ key: 'hud' })
  }

  preload () {
    this.load.setBaseURL('/res/')
    this.load.image('menu', 'img/textures/hud/menu.png')
    preloadMenuBorder(this.load)
  }

  create () {
    initMovementInput(this)
    initInteractInput(this)
    initTouchControls(this)
    initHudShortcutButtons(this)
    initMenuShortcuts(this)

    const btnMenu = this.add.image(60, 60, 'menu')
    btnMenu.setScrollFactor(0)
    btnMenu.setDepth(100)
    btnMenu.setInteractive({ useHandCursor: true })
    btnMenu.on('pointerdown', this.mnuClicked)

    this.floatingMessage = createFloatingMessageBox(this)
    this.hotbar = new Hotbar(this)

    this.input.keyboard?.on('keydown-ESC', this.onEscape, this)
    this.events.once('shutdown', this.onShutdown, this)

    window.hudScene = this
  }

  update () {
    this.floatingMessage?.update()
    this.hotbar?.refresh()
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
      sticky: msg.sticky === true,
      entityId: msg.entityId != null ? Number(msg.entityId) : null,
      worldX: msg.worldX,
      worldY: msg.worldY
    })
  }

  mnuClicked () {
    window.menuSystem?.toggle('main')
  }

  // Escape with no menu open opens the main menu. With a menu open, the menu
  // handles it as "back" (MenuSystem.onMenuKeyDown); it marks the event in
  // case it closed the menu before this listener ran.
  onEscape (event) {
    if (window.gameConsole?.isOpen?.() || event?.greyvarMenuHandled || window.menuSystem?.isOpen?.()) {
      return
    }
    window.menuSystem?.open('main')
  }

  onShutdown () {
    this.input.keyboard?.off('keydown-ESC', this.onEscape, this)
    this.floatingMessage?.destroy()
    this.floatingMessage = null
    this.hotbar?.destroy()
    this.hotbar = null
    if (window.hudScene === this) {
      window.hudScene = null
    }
  }
}
