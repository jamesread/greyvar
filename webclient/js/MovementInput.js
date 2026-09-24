import Phaser from 'phaser'

let keyboard = null
let cursors = null
let wasd = null
let lastMoveAt = 0

const MOVE_INTERVAL_MS = 67

function bindKeys (scene) {
  if (scene?.input?.keyboard == null) {
    return false
  }

  keyboard = scene.input.keyboard
  keyboard.enabled = true
  keyboard.addCapture([
    'W', 'A', 'S', 'D',
    'UP', 'DOWN', 'LEFT', 'RIGHT'
  ])
  cursors = keyboard.createCursorKeys()
  wasd = keyboard.addKeys({
    W: Phaser.Input.Keyboard.KeyCodes.W,
    A: Phaser.Input.Keyboard.KeyCodes.A,
    S: Phaser.Input.Keyboard.KeyCodes.S,
    D: Phaser.Input.Keyboard.KeyCodes.D
  })
  return true
}

export function initMovementInput (scene) {
  bindKeys(scene)
}

export function pollMovementInput () {
  if (keyboard == null || cursors == null || wasd == null || !window.serverConnection?.isOk) {
    const grid = window.phaser?.scene?.getScene?.('grid')
    if (grid != null) {
      bindKeys(grid)
    }
    if (keyboard == null || cursors == null || wasd == null || !window.serverConnection?.isOk) {
      return
    }
  }

  if (!window.phaser?.scene?.isActive('grid')) {
    return
  }

  if (window.menuSystem?.isOpen()) {
    return
  }

  if (window.gameConsole?.isOpen()) {
    return
  }

  // Grid scene owns gameplay input after teleports / scene restarts.
  if (keyboard.enabled === false) {
    keyboard.enabled = true
  }

  let x = 0
  let y = 0

  if (cursors.left.isDown || wasd.A.isDown) {
    x -= 1
  }
  if (cursors.right.isDown || wasd.D.isDown) {
    x += 1
  }
  if (cursors.up.isDown || wasd.W.isDown) {
    y -= 1
  }
  if (cursors.down.isDown || wasd.S.isDown) {
    y += 1
  }

  if (x === 0 && y === 0) {
    return
  }

  const now = performance.now()
  if (now - lastMoveAt < MOVE_INTERVAL_MS) {
    return
  }
  lastMoveAt = now

  window.serverConnection.sendMoveRequest({ x, y })
}

export function isMovementKeyDown () {
  if (keyboard == null || cursors == null || wasd == null) {
    return false
  }

  return (
    cursors.left.isDown ||
    cursors.right.isDown ||
    cursors.up.isDown ||
    cursors.down.isDown ||
    wasd.A.isDown ||
    wasd.D.isDown ||
    wasd.W.isDown ||
    wasd.S.isDown
  )
}
