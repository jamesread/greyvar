import { MENU_THEME } from './menu/MenuTheme.js'

const BUTTON_SIZE = 44
const BUTTON_GAP = 4
const MARGIN = 20

/** How far up from the bottom edge the on-screen D-pad reaches. */
export const TOUCH_CONTROLS_HEIGHT = MARGIN + 3 * BUTTON_SIZE + 2 * BUTTON_GAP
const DEPTH = 99

const directions = {
  up: false,
  down: false,
  left: false,
  right: false
}

let container = null
let buttons = {}
let sceneRef = null

export function isTouchDevice () {
  return (
    window.matchMedia('(pointer: coarse)').matches ||
    navigator.maxTouchPoints > 0
  )
}

function setDirection (dir, active) {
  directions[dir] = active
  const btn = buttons[dir]
  if (btn != null) {
    btn.setAlpha(active ? 1 : 0.75)
  }
}

function clearDirections () {
  for (const dir of Object.keys(directions)) {
    setDirection(dir, false)
  }
}

function layoutControls () {
  if (container == null || sceneRef == null) {
    return
  }

  const cx = MARGIN + BUTTON_SIZE * 1.5 + BUTTON_GAP
  const cy = sceneRef.scale.height - MARGIN - BUTTON_SIZE * 1.5 - BUTTON_GAP
  container.setPosition(cx, cy)

  const step = BUTTON_SIZE + BUTTON_GAP

  buttons.up.setPosition(0, -step)
  buttons.down.setPosition(0, step)
  buttons.left.setPosition(-step, 0)
  buttons.right.setPosition(step, 0)
}

function makeButton (scene, dir) {
  const btn = scene.add.rectangle(0, 0, BUTTON_SIZE, BUTTON_SIZE, MENU_THEME.panel, 0.75)
  btn.setStrokeStyle(2, MENU_THEME.panelBorder, 0.9)
  btn.setScrollFactor(0)
  btn.setDepth(DEPTH)
  btn.setInteractive({ useHandCursor: false })

  btn.on('pointerdown', () => setDirection(dir, true))
  btn.on('pointerup', () => setDirection(dir, false))
  btn.on('pointerout', () => setDirection(dir, false))
  btn.on('pointerover', (pointer) => {
    if (pointer.isDown) {
      setDirection(dir, true)
    }
  })

  return btn
}

export function initTouchControls (scene) {
  if (!isTouchDevice() || container != null) {
    return
  }

  sceneRef = scene
  container = scene.add.container(0, 0)
  container.setScrollFactor(0)
  container.setDepth(DEPTH)

  buttons = {
    up: makeButton(scene, 'up'),
    down: makeButton(scene, 'down'),
    left: makeButton(scene, 'left'),
    right: makeButton(scene, 'right')
  }

  for (const btn of Object.values(buttons)) {
    container.add(btn)
  }

  scene.input.on('pointerup', clearDirections)
  scene.scale.on('resize', layoutControls)
  scene.events.once('shutdown', destroyTouchControls)

  layoutControls()
}

export function destroyTouchControls () {
  if (sceneRef != null) {
    sceneRef.input?.off('pointerup', clearDirections)
    sceneRef.scale?.off('resize', layoutControls)
  }

  clearDirections()
  container?.destroy(true)
  container = null
  buttons = {}
  sceneRef = null
}

export function getTouchMovementDirection () {
  let x = 0
  let y = 0

  if (directions.left) {
    x -= 1
  }
  if (directions.right) {
    x += 1
  }
  if (directions.up) {
    y -= 1
  }
  if (directions.down) {
    y += 1
  }

  return { x, y }
}

export function isTouchMovementActive () {
  return directions.up || directions.down || directions.left || directions.right
}
