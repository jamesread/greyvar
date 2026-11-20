import Phaser from 'phaser'
import { facedCell } from './facedCell.js'

let interactKey = null

export function initInteractInput (scene) {
  if (scene?.input?.keyboard == null) {
    return false
  }

  interactKey = scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.E)
  return true
}

/**
 * The entity the player means: among those within the server's interact
 * range, the one nearest the point just ahead of them (facing), so pressing E
 * towards an NPC talks to it even when a sign beside you is slightly closer.
 * rangeFor(ent) may widen an entity's reach (entdef interactRange, e.g. an
 * NPC behind a counter); it returns 0/undefined for the default.
 */
export function pickInteractTarget (entities, localId, lx, ly, facing = 'Down', rangeFor = () => 0) {
  const f = String(facing ?? 'Down')
  const ax = lx + (f.includes('Left') ? -AHEAD_PX : f.includes('Right') ? AHEAD_PX : 0)
  const ay = ly + (f.includes('Up') ? -AHEAD_PX : f.includes('Down') ? AHEAD_PX : 0)

  let best = null
  let bestScore = Infinity
  for (const ent of Object.values(entities)) {
    if (ent.entityId === localId) {
      continue
    }
    const ex = ent.x ?? ent.img?.x
    const ey = ent.y ?? ent.img?.y
    if (ex == null || ey == null) {
      continue
    }
    // In reach as the server measures it (24px or the entdef's
    // interactRange, Chebyshev, from the player); the point ahead only
    // decides between candidates.
    if (Math.max(Math.abs(ex - lx), Math.abs(ey - ly)) > (rangeFor(ent) || RANGE_PX)) {
      continue
    }
    const ahead = Math.max(Math.abs(ex - ax), Math.abs(ey - ay))
    if (ahead < bestScore) {
      best = ent
      bestScore = ahead
    }
  }
  return best
}

const RANGE_PX = 24
const AHEAD_PX = 12

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
  const local = localId != null ? grid.entities[localId] : null
  if (local == null) {
    return
  }
  // Send the nearest entity (if any) and the faced cell; the server uses the
  // entity when it has something to do, otherwise the cell (plant/harvest).
  const facing = local.img?.facing ?? 'Down'
  // Play the use animation straight away; the server shows it to others.
  local.img?.playUse?.(facing)
  const lx = local.x ?? local.img?.x ?? 0
  const ly = local.y ?? local.img?.y ?? 0
  const rangeFor = (ent) => window.gameState?.entdefs?.get(ent.definition)?.interactRange ?? 0
  const target = pickInteractTarget(grid.entities, localId, lx, ly, facing, rangeFor)
  const cell = facedCell(local.x ?? local.img?.x ?? 0, local.y ?? local.img?.y ?? 0, facing)
  window.serverConnection.sendInteractRequest(target?.entityId ?? 0, cell, facing)
}
