const COLLISION_DEBUG_STORAGE_KEY = 'greyvar.debug.collision'

const DEFAULT_BLOCKING_TEXTURES = [
  'water.png',
  'barrier.png',
  'fence.png'
]

let blockingTextures = new Set(DEFAULT_BLOCKING_TEXTURES)
let blockingTexturesLoaded = false
let blockingTexturesLoading = null

const TILE_SIZE = 16
const ENTITY_SIZE = 16

export async function ensureBlockingTextures () {
  if (blockingTexturesLoaded) {
    return blockingTextures
  }

  if (blockingTexturesLoading != null) {
    return blockingTexturesLoading
  }

  blockingTexturesLoading = fetch('/api/debug/blocking-tiles')
    .then(async (res) => {
      if (!res.ok) {
        throw new Error(`${res.status} ${res.statusText}`)
      }
      return res.json()
    })
    .then((data) => {
      const textures = data?.textures
      if (Array.isArray(textures) && textures.length > 0) {
        blockingTextures = new Set(textures)
      }
      blockingTexturesLoaded = true
      return blockingTextures
    })
    .catch(() => {
      blockingTextures = new Set(DEFAULT_BLOCKING_TEXTURES)
      blockingTexturesLoaded = true
      return blockingTextures
    })
    .finally(() => {
      blockingTexturesLoading = null
    })

  return blockingTexturesLoading
}

export function isLegacyTileBlocked (tile) {
  if (tile == null || tile.textureName == null || tile.textureName === '') {
    return false
  }

  return blockingTextures.has(tile.textureName)
}

export function transformCollisionRect (rect, tileSize, flipH, flipV) {
  let x = rect.x
  let y = rect.y
  const w = rect.w
  const h = rect.h

  if (flipH) {
    x = tileSize - rect.x - w
  }
  if (flipV) {
    y = tileSize - rect.y - h
  }

  return { x, y, w, h }
}

export function tileCollisionRectsWorld (tile) {
  if (tile?.collision == null || tile.collision.length === 0) {
    return []
  }

  const originX = tile.col * TILE_SIZE
  const originY = tile.row * TILE_SIZE

  return tile.collision.map((rect) => {
    const local = transformCollisionRect(
      rect,
      TILE_SIZE,
      tile.textureHorizontalFlip,
      tile.textureVerticalFlip
    )
    return {
      x: originX + local.x,
      y: originY + local.y,
      w: local.w,
      h: local.h
    }
  })
}

/** Entity-state collision rects in world pixels (tile-local shapes from entity tileset). */
export function entityCollisionRectsWorld (ent, entdef) {
  if (ent == null || entdef == null || !Array.isArray(entdef.states)) {
    return []
  }

  const stateName = ent.state || ent.initialState
  const state = entdef.states.find((s) => s?.name === stateName)
    || entdef.states.find((s) => s?.name === ent.initialState)
    || entdef.states[0]
  if (state?.collision == null || state.collision.length === 0) {
    return []
  }

  const originX = ent.img?.x ?? ent.x ?? 0
  const originY = ent.img?.y ?? ent.y ?? 0

  return state.collision.map((rect) => ({
    x: originX + (rect.x ?? 0),
    y: originY + (rect.y ?? 0),
    w: rect.w ?? 0,
    h: rect.h ?? 0
  }))
}

export function collisionDebugStats (grid) {
  let tsxRectCount = 0
  let legacyBlockedCells = 0
  let passableCells = 0

  for (let row = 0; row < grid.rowCount; row++) {
    for (let col = 0; col < grid.colCount; col++) {
      const tile = grid.topTileAt(row, col)
      if (tile?.collision?.length > 0) {
        tsxRectCount += tile.collision.length
      } else if (tile?.collisionFromTsx) {
        passableCells++
      } else if (isLegacyTileBlocked(tile)) {
        legacyBlockedCells++
      } else {
        passableCells++
      }
    }
  }

  return { tsxRectCount, legacyBlockedCells, passableCells }
}

export function isCollisionDebugEnabled () {
  return window.gameState?.collisionDebugEnabled === true
}

function readStoredCollisionDebugEnabled () {
  try {
    return localStorage.getItem(COLLISION_DEBUG_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

function writeStoredCollisionDebugEnabled (enabled) {
  try {
    if (enabled) {
      localStorage.setItem(COLLISION_DEBUG_STORAGE_KEY, 'true')
    } else {
      localStorage.removeItem(COLLISION_DEBUG_STORAGE_KEY)
    }
  } catch {
    // ignore quota / private browsing errors
  }
}

export function initCollisionDebugPreference (gameState) {
  if (gameState == null) {
    return
  }

  gameState.collisionDebugEnabled = readStoredCollisionDebugEnabled()
}

function setCollisionDebugEnabled (enabled) {
  if (window.gameState == null) {
    return
  }

  window.gameState.collisionDebugEnabled = enabled
  writeStoredCollisionDebugEnabled(enabled)
}

export async function refreshCollisionDebugOverlay () {
  const scene = window.gameState?.gridScene
  if (scene == null) {
    return
  }

  if (!isCollisionDebugEnabled()) {
    scene.hideCollisionDebugOverlay()
    return
  }

  await ensureBlockingTextures()
  scene.showCollisionDebugOverlay()
}

export async function toggleCollisionDebug () {
  if (window.gameState?.gridScene == null) {
    return 'grid not loaded'
  }

  setCollisionDebugEnabled(!isCollisionDebugEnabled())
  await refreshCollisionDebugOverlay()

  if (!isCollisionDebugEnabled()) {
    return 'collision debug overlay disabled'
  }

  const grid = window.gameState.grid
  const stats = grid != null ? collisionDebugStats(grid) : null

  return [
    'collision debug overlay enabled',
    'red = tile blocking, orange = entity tileset collision, blue = player',
    'green = passable whole tile (legacy grids)',
    'blue outline = player entity bounds (16x16)',
    stats != null
      ? `tsx collision rects: ${stats.tsxRectCount}, legacy blocked cells: ${stats.legacyBlockedCells}, passable cells: ${stats.passableCells}`
      : null
  ].filter(Boolean)
}

export { TILE_SIZE, ENTITY_SIZE }
