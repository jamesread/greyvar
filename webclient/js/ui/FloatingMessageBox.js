/**
 * Reusable floating message box for world-anchored prompts (signs, pickups, etc.).
 * Drawn in a HUD/overlay scene with scrollFactor 0; positions from world→screen.
 */

const DEFAULTS = {
  fontFamily: 'Helvetica Neue, Helvetica, Arial, sans-serif',
  fontSize: '16px',
  color: '#f2f2f2',
  backgroundColor: 0x4a4a4a,
  backgroundAlpha: 0.92,
  paddingX: 12,
  paddingY: 8,
  cornerRadius: 10,
  maxWidth: 280,
  gap: 10,
  screenPadding: 12,
  depth: 120
}

/**
 * @param {Phaser.Scene} scene
 * @param {object} [style]
 */
export function createFloatingMessageBox (scene, style = {}) {
  const opts = { ...DEFAULTS, ...style }

  const root = scene.add.container(0, 0)
  root.setScrollFactor(0)
  root.setDepth(opts.depth)
  root.setVisible(false)

  const bg = scene.add.graphics()
  const label = scene.add.text(0, 0, '', {
    fontFamily: opts.fontFamily,
    fontSize: opts.fontSize,
    color: opts.color,
    align: 'left',
    wordWrap: { width: opts.maxWidth }
  })
  label.setOrigin(0, 0)

  root.add([bg, label])

  /** @type {null | { text: string, entityId?: number|null, worldX?: number|null, worldY?: number|null }} */
  let active = null

  function measure () {
    const tw = Math.ceil(label.width)
    const th = Math.ceil(label.height)
    const w = tw + opts.paddingX * 2
    const h = th + opts.paddingY * 2
    return { tw, th, w, h }
  }

  function redrawBackground () {
    const { tw, th, w, h } = measure()
    bg.clear()
    bg.fillStyle(opts.backgroundColor, opts.backgroundAlpha)
    bg.fillRoundedRect(0, 0, w, h, opts.cornerRadius)
    label.setPosition(opts.paddingX, opts.paddingY)
    return { w, h, tw, th }
  }

  function hide () {
    active = null
    label.setText('')
    bg.clear()
    root.setVisible(false)
  }

  /**
   * Show or update a floating message.
   * @param {{ text: string, entityId?: number|null, worldX?: number|null, worldY?: number|null }} msg
   */
  function show (msg) {
    const text = typeof msg?.text === 'string' ? msg.text.trim() : ''
    if (text === '') {
      hide()
      return
    }
    active = {
      text,
      entityId: msg.entityId ?? null,
      worldX: msg.worldX ?? null,
      worldY: msg.worldY ?? null
    }
    label.setWordWrapWidth(opts.maxWidth)
    label.setText(text)
    redrawBackground()
    root.setVisible(true)
    layout()
  }

  function resolveAnchorWorld () {
    if (active == null) {
      return null
    }

    const gridScene = window.gameState?.gridScene
    if (active.entityId != null && active.entityId !== 0 && gridScene?.entities) {
      const ent = gridScene.entities[active.entityId]
        ?? gridScene.entities[String(active.entityId)]
      if (ent?.img != null) {
        const displayW = ent.img.displayWidth || 16
        const displayH = ent.img.displayHeight || 16
        return {
          x: ent.img.x + displayW / 2,
          y: ent.img.y + displayH / 2,
          w: displayW,
          h: displayH,
          left: ent.img.x,
          top: ent.img.y,
          right: ent.img.x + displayW,
          bottom: ent.img.y + displayH
        }
      }
      if (ent != null && ent.x != null) {
        return {
          x: ent.x + 8,
          y: ent.y + 8,
          w: 16,
          h: 16,
          left: ent.x,
          top: ent.y,
          right: ent.x + 16,
          bottom: ent.y + 16
        }
      }
    }

    if (active.worldX != null && active.worldY != null) {
      const x = active.worldX
      const y = active.worldY
      return { x, y, w: 0, h: 0, left: x, top: y, right: x, bottom: y }
    }

    return null
  }

  function playerScreenRect (cam, viewW, viewH) {
    const gridScene = window.gameState?.gridScene
    const localId = window.gameState?.localPlayerEntityId
    const ent = localId != null ? gridScene?.entities?.[localId] : null
    if (ent?.img == null || cam == null) {
      return null
    }
    const dw = ent.img.displayWidth || 16
    const dh = ent.img.displayHeight || 16
    const tl = worldToScreen(cam, ent.img.x, ent.img.y, viewW, viewH)
    const br = worldToScreen(cam, ent.img.x + dw, ent.img.y + dh, viewW, viewH)
    return {
      left: Math.min(tl.x, br.x),
      top: Math.min(tl.y, br.y),
      right: Math.max(tl.x, br.x),
      bottom: Math.max(tl.y, br.y)
    }
  }

  function worldToScreen (cam, wx, wy, viewW, viewH) {
    const wv = cam.worldView
    return {
      x: ((wx - wv.x) / wv.width) * viewW,
      y: ((wy - wv.y) / wv.height) * viewH
    }
  }

  function rectsOverlap (a, b) {
    return a.left < b.right && a.right > b.left && a.top < b.bottom && a.bottom > b.top
  }

  function clampBox (x, y, w, h, viewW, viewH) {
    const pad = opts.screenPadding
    let nx = x
    let ny = y
    if (nx < pad) nx = pad
    if (ny < pad) ny = pad
    if (nx + w > viewW - pad) nx = Math.max(pad, viewW - pad - w)
    if (ny + h > viewH - pad) ny = Math.max(pad, viewH - pad - h)
    return { x: nx, y: ny }
  }

  function layout () {
    if (active == null || !root.visible) {
      return
    }

    const { w, h } = redrawBackground()
    const viewW = scene.scale.width
    const viewH = scene.scale.height
    const gridScene = window.gameState?.gridScene
    const cam = gridScene?.cameras?.main

    const anchor = resolveAnchorWorld()
    if (anchor == null || cam == null) {
      // Fallback: bottom-center of screen
      const pos = clampBox((viewW - w) / 2, viewH - h - opts.screenPadding, w, h, viewW, viewH)
      root.setPosition(pos.x, pos.y)
      return
    }

    const gap = opts.gap
    const aTL = worldToScreen(cam, anchor.left, anchor.top, viewW, viewH)
    const aBR = worldToScreen(cam, anchor.right, anchor.bottom, viewW, viewH)
    const aLeft = Math.min(aTL.x, aBR.x)
    const aTop = Math.min(aTL.y, aBR.y)
    const aRight = Math.max(aTL.x, aBR.x)
    const aBottom = Math.max(aTL.y, aBR.y)
    const aCx = (aLeft + aRight) / 2
    const aCy = (aTop + aBottom) / 2

    const player = playerScreenRect(cam, viewW, viewH)

    const candidates = [
      { x: aCx - w / 2, y: aTop - gap - h }, // above
      { x: aRight + gap, y: aCy - h / 2 }, // right
      { x: aLeft - gap - w, y: aCy - h / 2 }, // left
      { x: aCx - w / 2, y: aBottom + gap } // below
    ]

    let best = null
    let bestScore = -Infinity

    for (const c of candidates) {
      const clamped = clampBox(c.x, c.y, w, h, viewW, viewH)
      const box = {
        left: clamped.x,
        top: clamped.y,
        right: clamped.x + w,
        bottom: clamped.y + h
      }

      // Prefer placements that stay near the preferred candidate point.
      const drift = Math.hypot(clamped.x - c.x, clamped.y - c.y)
      let score = 1000 - drift

      if (player != null && rectsOverlap(box, player)) {
        score -= 500
      }

      // Prefer not covering the anchor itself.
      const anchorBox = { left: aLeft, top: aTop, right: aRight, bottom: aBottom }
      if (rectsOverlap(box, anchorBox)) {
        score -= 50
      }

      if (score > bestScore) {
        bestScore = score
        best = clamped
      }
    }

    if (best == null) {
      best = clampBox((viewW - w) / 2, viewH - h - opts.screenPadding, w, h, viewW, viewH)
    }

    // If still overlapping player, nudge away from player center.
    if (player != null) {
      const box = {
        left: best.x,
        top: best.y,
        right: best.x + w,
        bottom: best.y + h
      }
      if (rectsOverlap(box, player)) {
        const pcx = (player.left + player.right) / 2
        const pcy = (player.top + player.bottom) / 2
        const bcx = best.x + w / 2
        const bcy = best.y + h / 2
        const dx = bcx - pcx
        const dy = bcy - pcy
        const nudge = 24
        if (Math.abs(dx) >= Math.abs(dy)) {
          best.x += dx >= 0 ? nudge : -nudge
        } else {
          best.y += dy >= 0 ? nudge : -nudge
        }
        best = clampBox(best.x, best.y, w, h, viewW, viewH)
      }
    }

    root.setPosition(best.x, best.y)
  }

  function update () {
    if (active == null || !root.visible) {
      return
    }
    layout()
  }

  function destroy () {
    hide()
    root.destroy(true)
  }

  return {
    show,
    hide,
    update,
    layout,
    destroy,
    get root () { return root },
    isVisible () { return root.visible },
    getActive () { return active }
  }
}
