// `debug pos`: the local player's position, bottom-left of the screen.

const KEY = 'greyvar.debugPos'
const TILE = 16
let enabled = readStored()
let el = null
let lastText = ''

function readStored () {
  try {
    return globalThis.localStorage?.getItem(KEY) === '1'
  } catch {
    return false
  }
}

function writeStored (on) {
  try {
    globalThis.localStorage?.setItem(KEY, on ? '1' : '0')
  } catch {
    // Blocked storage: the toggle just won't be remembered.
  }
}

/** One line: pixel position, tile (row,col) and map. */
export function formatDebugPos (ent, grid) {
  if (ent == null) {
    return 'pos: no player'
  }
  const x = Math.round(ent.x ?? 0)
  const y = Math.round(ent.y ?? 0)
  // Tile under the player's 16px cell centre.
  const row = Math.floor((y + TILE / 2) / TILE)
  const col = Math.floor((x + TILE / 2) / TILE)
  const where = [grid?.worldId, grid?.gridId].filter(Boolean).join(' / ')
  return `x ${x}  y ${y}  |  tile row ${row} col ${col}${where ? '  |  ' + where : ''}`
}

export function isDebugPosEnabled () {
  return enabled
}

export function toggleDebugPos () {
  enabled = !enabled
  writeStored(enabled)
  if (!enabled) {
    el?.remove()
    el = null
    lastText = ''
  }
  return `debug pos ${enabled ? 'on' : 'off'}`
}

/** Called every frame by GridScene. */
export function updateDebugPos (scene) {
  if (!enabled || typeof document === 'undefined') {
    return
  }
  if (el == null) {
    el = document.createElement('div')
    el.id = 'greyvar-debug-pos'
    Object.assign(el.style, {
      position: 'fixed',
      left: '8px',
      bottom: '8px',
      padding: '4px 8px',
      background: 'rgba(20, 18, 24, 0.75)',
      color: '#e8f0e0',
      font: '12px monospace',
      borderRadius: '4px',
      zIndex: 1000,
      pointerEvents: 'none'
    })
    document.body.appendChild(el)
  }
  const id = window.gameState?.localPlayerEntityId
  const ent = id != null ? scene.entities?.[id] : null
  // Show where the sprite is (smooth while walking), not just the last server step.
  const shown = ent?.img != null ? { x: ent.img.x, y: ent.img.y } : ent
  const text = formatDebugPos(shown, scene.grid)
  if (text !== lastText) {
    el.textContent = text
    lastText = text
  }
}
