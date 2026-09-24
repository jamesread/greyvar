import { MENU_THEME } from './MenuTheme.js'

const DEFAULT_MENU_ID = 'main'

export default class MenuSystem {
  constructor (game) {
    this.game = game
    this.menus = new Map()
    this.stack = []
    this.sceneKey = 'menu'
    this.boundScene = null
    this.ui = null
    this.focusIndex = 0
    this.focusItemId = null
  }

  static init (game) {
    if (window.menuSystem == null) {
      window.menuSystem = new MenuSystem(game)
    }
    return window.menuSystem
  }

  registerMenu (config) {
    if (config?.id == null) {
      throw new Error('MenuSystem.registerMenu requires an id')
    }

    this.menus.set(config.id, {
      id: config.id,
      title: config.title ?? config.id,
      subtitle: config.subtitle ?? '',
      items: config.items ?? []
    })

    return this
  }

  getMenu (menuId) {
    return this.menus.get(menuId)
  }

  currentMenuId () {
    return this.stack[this.stack.length - 1] ?? null
  }

  currentMenu () {
    const id = this.currentMenuId()
    return id == null ? null : this.getMenu(id)
  }

  isOpen () {
    return this.game.scene.isActive(this.sceneKey)
  }

  open (menuId = DEFAULT_MENU_ID) {
    if (!this.menus.has(menuId)) {
      console.warn('[MenuSystem] unknown menu:', menuId)
      return
    }

    this.stack = [menuId]
    this.focusIndex = 0
    this.focusItemId = null
    this.ensureSceneRunning()
    this.render()
  }

  close () {
    this.stack = []
    this.focusIndex = 0
    this.focusItemId = null
    this.clearUi()

    if (this.game.scene.isActive(this.sceneKey)) {
      this.game.scene.stop(this.sceneKey)
    }
  }

  toggle (menuId = DEFAULT_MENU_ID) {
    if (this.isOpen()) {
      this.close()
    } else {
      this.open(menuId)
    }
  }

  navigateTo (menuId) {
    if (!this.menus.has(menuId)) {
      console.warn('[MenuSystem] unknown menu:', menuId)
      return
    }

    this.stack.push(menuId)
    this.focusIndex = 0
    this.focusItemId = null
    this.render()
  }

  navigateBack () {
    if (this.stack.length <= 1) {
      this.close()
      return
    }

    this.stack.pop()
    this.focusIndex = 0
    this.focusItemId = null
    this.render()
  }

  ensureSceneRunning () {
    const mgr = this.game.scene
    if (!mgr.getScene(this.sceneKey)) {
      return
    }

    if (!mgr.isActive(this.sceneKey)) {
      mgr.run(this.sceneKey)
    }

    mgr.bringToTop(this.sceneKey)
    if (mgr.getScene('hud')) {
      mgr.bringToTop('hud')
    }
  }

  attachScene (scene) {
    this.boundScene = scene
    this.bindKeyboard(scene)
    if (this.isOpen()) {
      this.render()
    }
  }

  detachScene () {
    this.unbindKeyboard()
    this.clearUi()
    this.boundScene = null
  }

  bindKeyboard (scene) {
    this.unbindKeyboard()
    const keyboard = scene.input.keyboard
    if (keyboard == null) {
      return
    }

    this._onMenuKeyDown = (event) => this.onMenuKeyDown(event)
    keyboard.on('keydown', this._onMenuKeyDown)
  }

  unbindKeyboard () {
    const keyboard = this.boundScene?.input?.keyboard
    if (keyboard != null && this._onMenuKeyDown != null) {
      keyboard.off('keydown', this._onMenuKeyDown)
    }
    this._onMenuKeyDown = null
  }

  onMenuKeyDown (event) {
    if (!this.isOpen() || this.ui?.itemRows == null) {
      return
    }

    if (window.gameConsole?.isOpen?.()) {
      return
    }

    const code = event.code ?? event.keyCode
    const key = event.key

    if (code === 'ArrowDown' || code === 'KeyS' || key === 'ArrowDown' || key === 's' || key === 'S') {
      event.preventDefault?.()
      this.moveFocus(1)
      return
    }

    if (code === 'ArrowUp' || code === 'KeyW' || key === 'ArrowUp' || key === 'w' || key === 'W') {
      event.preventDefault?.()
      this.moveFocus(-1)
      return
    }

    if (code === 'Enter' || code === 'NumpadEnter' || code === 'Space' || key === 'Enter' || key === ' ') {
      event.preventDefault?.()
      this.activateFocused()
      return
    }

    if (code === 'ArrowLeft' || code === 'KeyA' || code === 'Backspace' ||
        key === 'ArrowLeft' || key === 'a' || key === 'A' || key === 'Backspace') {
      if (this.stack.length > 1) {
        event.preventDefault?.()
        this.navigateBack()
      }
      return
    }

    if (code === 'ArrowRight' || code === 'KeyD' || key === 'ArrowRight' || key === 'd' || key === 'D') {
      event.preventDefault?.()
      this.activateFocused()
    }
  }

  selectableRows () {
    return (this.ui?.itemRows ?? []).filter((row) => row.enabled)
  }

  moveFocus (delta) {
    const rows = this.selectableRows()
    if (rows.length === 0) {
      return
    }

    const currentPos = rows.findIndex((row) => row.index === this.focusIndex)
    const from = currentPos >= 0 ? currentPos : 0
    const next = (from + delta + rows.length * 10) % rows.length
    this.setFocus(rows[next].index)
  }

  setFocus (index) {
    const rows = this.ui?.itemRows
    if (rows == null || rows.length === 0) {
      return
    }

    const clamped = Math.max(0, Math.min(index, rows.length - 1))
    this.focusIndex = clamped
    this.focusItemId = rows[clamped]?.id ?? null

    for (const row of rows) {
      this.applyFocusStyle(row, row.index === clamped)
    }
  }

  activateFocused () {
    const row = this.ui?.itemRows?.[this.focusIndex]
    if (row == null || !row.enabled) {
      return
    }
    row.activate()
  }

  clearUi () {
    if (this.ui?.root != null) {
      this.boundScene?.scale?.off('resize', this.ui.layoutHandler)
      this.ui.root.destroy(true)
    }
    this.ui = null
  }

  resolveValue (value, ctx) {
    if (typeof value === 'function') {
      return value(ctx)
    }
    return value ?? true
  }

  applyFocusStyle (row, focused) {
    if (!row.enabled) {
      row.focusBg.setVisible(false)
      row.caret.setVisible(false)
      row.label.setColor(MENU_THEME.itemDisabled)
      return
    }

    row.focusBg.setVisible(focused)
    row.caret.setVisible(focused)
    row.label.setColor(focused ? MENU_THEME.itemFocus : MENU_THEME.item)
  }

  resolveInitialFocus (itemRows) {
    if (this.focusItemId != null) {
      const byId = itemRows.findIndex((row) => row.id === this.focusItemId && row.enabled)
      if (byId >= 0) {
        return byId
      }
    }

    if (itemRows[this.focusIndex]?.enabled) {
      return this.focusIndex
    }

    const firstEnabled = itemRows.findIndex((row) => row.enabled)
    return firstEnabled >= 0 ? firstEnabled : 0
  }

  render () {
    const scene = this.boundScene
    const menu = this.currentMenu()

    if (scene == null || menu == null) {
      return
    }

    this.clearUi()

    const root = scene.add.container(0, 0)
    root.setScrollFactor(0)
    root.setDepth(1000)

    const overlay = scene.add.rectangle(0, 0, 10, 10, MENU_THEME.bgOverlay, MENU_THEME.bgOverlayAlpha)
    overlay.setOrigin(0)
    overlay.setInteractive()
    overlay.on('pointerdown', () => this.close())

    const panelGfx = scene.add.graphics()
    const title = scene.add.text(0, 0, menu.title, {
      fontFamily: MENU_THEME.fontFamily,
      fontSize: '28px',
      color: MENU_THEME.title,
      fontStyle: 'bold'
    }).setOrigin(0.5, 0)

    const subtitle = menu.subtitle
      ? scene.add.text(0, 0, menu.subtitle, {
        fontFamily: MENU_THEME.fontFamily,
        fontSize: '13px',
        color: MENU_THEME.subtitle
      }).setOrigin(0.5, 0)
      : null

    const itemRows = []
    const ctx = {
      menuSystem: this,
      scene,
      game: this.game
    }

    if (this.stack.length > 1) {
      itemRows.push(this.makeItem(scene, {
        id: '__back',
        label: 'Back',
        action: () => this.navigateBack()
      }, ctx, itemRows.length))
    }

    for (const item of menu.items) {
      if (!this.resolveValue(item.visible ?? true, ctx)) {
        continue
      }

      itemRows.push(this.makeItem(scene, item, ctx, itemRows.length))
    }

    root.add([overlay, panelGfx, title])
    if (subtitle != null) {
      root.add(subtitle)
    }
    for (const row of itemRows) {
      root.add(row.container)
    }

    const layout = () => {
      const { width, height } = scene.scale
      const pad = 28
      const panelW = Math.min(360, width - 48)
      const rowH = 40
      const headerH = subtitle != null ? 88 : 64
      const panelH = pad * 2 + headerH + itemRows.length * rowH
      const panelX = (width - panelW) / 2
      const panelY = Math.max(24, (height - panelH) / 2)
      const innerW = panelW - pad * 2

      overlay.setPosition(0, 0)
      overlay.setSize(width, height)

      panelGfx.clear()
      panelGfx.fillStyle(MENU_THEME.panel, 1)
      panelGfx.fillRoundedRect(panelX, panelY, panelW, panelH, 8)
      panelGfx.lineStyle(2, MENU_THEME.panelBorderDark, 1)
      panelGfx.strokeRoundedRect(panelX, panelY, panelW, panelH, 8)
      panelGfx.lineStyle(1, MENU_THEME.panelBorder, 0.9)
      panelGfx.strokeRoundedRect(panelX + 3, panelY + 3, panelW - 6, panelH - 6, 6)

      title.setPosition(width / 2, panelY + pad + 4)
      if (subtitle != null) {
        subtitle.setPosition(width / 2, panelY + pad + 36)
      }

      let rowY = panelY + pad + headerH
      for (const row of itemRows) {
        row.container.setPosition(panelX + pad, rowY)
        row.hitArea.setSize(innerW, rowH - 4)
        row.focusBg.setSize(innerW, rowH - 4)
        row.label.setWordWrapWidth(innerW - 28)
        rowY += rowH
      }
    }

    layout()
    this.ui = {
      root,
      layout,
      layoutHandler: layout,
      itemRows
    }
    scene.scale.on('resize', layout)

    this.setFocus(this.resolveInitialFocus(itemRows))
  }

  makeItem (scene, item, ctx, index) {
    const container = scene.add.container(0, 0)
    const enabled = this.resolveValue(item.enabled ?? true, ctx)
    const labelText = this.resolveValue(item.label, ctx)
    const itemId = item.id ?? `item-${index}`

    const focusBg = scene.add.rectangle(0, 0, 10, 10, MENU_THEME.focusFill, MENU_THEME.focusFillAlpha)
    focusBg.setOrigin(0, 0)
    focusBg.setStrokeStyle(1, MENU_THEME.focusBorder, 1)
    focusBg.setVisible(false)

    const hitArea = scene.add.rectangle(0, 0, 10, 10, 0xffffff, 0.001)
    hitArea.setOrigin(0, 0)

    const caret = scene.add.text(6, 18, '▸', {
      fontFamily: MENU_THEME.fontFamily,
      fontSize: '16px',
      color: MENU_THEME.caret
    }).setOrigin(0, 0.5)
    caret.setVisible(false)

    const label = scene.add.text(22, 18, labelText, {
      fontFamily: MENU_THEME.fontFamily,
      fontSize: '18px',
      color: enabled ? MENU_THEME.item : MENU_THEME.itemDisabled
    }).setOrigin(0, 0.5)

    container.add([focusBg, hitArea, caret, label])

    const activate = () => {
      if (!enabled) {
        return
      }

      if (item.submenu != null) {
        this.navigateTo(item.submenu)
        return
      }

      if (typeof item.action === 'function') {
        item.action(ctx)
      }
    }

    if (enabled) {
      hitArea.setInteractive({ useHandCursor: true })
      hitArea.on('pointerover', () => this.setFocus(index))
      hitArea.on('pointerdown', (pointer) => {
        pointer.event.stopPropagation()
        this.setFocus(index)
        activate()
      })
    }

    return {
      id: itemId,
      container,
      hitArea,
      focusBg,
      caret,
      label,
      index,
      enabled,
      activate
    }
  }
}

export function registerDefaultMenus (menuSystem) {
  menuSystem.registerMenu({
    id: 'main',
    title: 'Menu',
    subtitle: 'Greyvar',
    items: [
      {
        id: 'continue',
        label: 'Continue',
        action: ({ menuSystem: ms }) => ms.close()
      },
      {
        id: 'fullscreen',
        label: 'Toggle fullscreen',
        action: () => {
          if (!window.phaser.scale.fullscreen.active) {
            window.phaser.scale.startFullscreen()
          } else {
            window.phaser.scale.stopFullscreen()
          }
        }
      },
      {
        id: 'music',
        label: () => window.backgroundMusic?.isMuted()
          ? 'Unmute music'
          : 'Mute music',
        action: ({ menuSystem: ms }) => {
          window.backgroundMusic?.toggleMute()
          ms.render()
        }
      },
      {
        id: 'reload',
        label: 'Reload game',
        action: () => window.location.reload()
      },
      {
        id: 'debug',
        label: 'Debug: dump movement state',
        action: () => window.dumpMovementDebug?.()
      }
    ]
  })
}
