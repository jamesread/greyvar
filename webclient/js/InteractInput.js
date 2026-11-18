import Phaser from 'phaser'

let interactKey = null

export function initInteractInput (scene) {
  if (scene?.input?.keyboard == null) {
    return false
  }

  interactKey = scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.E)
  return true
}

function findNearestEntity (grid, localId) {
  const local = localId != null ? grid.entities[localId] : null
  if (local == null) {
    return null
  }

  const lx = local.x ?? local.img?.x ?? 0
  const ly = local.y ?? local.img?.y ?? 0
  const range = 24

  let best = null
  let bestDist = Infinity

  for (const ent of Object.values(grid.entities)) {
    if (ent.entityId === localId) {
      continue
    }

    const ex = ent.x ?? ent.img?.x
    const ey = ent.y ?? ent.img?.y
    if (ex == null || ey == null) {
      continue
    }

    const dist = Math.max(Math.abs(ex - lx), Math.abs(ey - ly))
    if (dist <= range && dist < bestDist) {
      best = ent
      bestDist = dist
    }
  }

  return best
}

export function pollInteractInput () {
  if (interactKey == null || !Phaser.Input.Keyboard.JustDown(interactKey)) {
    return
  }

  if (!window.serverConnection?.isOk) {
    return
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

  const grid = window.phaser.scene.getScene('grid')
  if (grid?.isScrollTransitioning) {
    return
  }

  const localId = window.gameState?.localPlayerEntityId
  const target = findNearestEntity(grid, localId)
  if (target?.entityId == null) {
    return
  }

  window.serverConnection.sendInteractRequest(target.entityId)
}
