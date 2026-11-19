import GridScene from './scenes/GridScene.js'
import { loadSavedAppearance, restoreRequest } from './sprites/appearanceStore.js'
import { normalizeEntitySpawn, normalizeEntityPosition, entityId } from './entityUtils.js'
import { logGreyvar, sceneSnapshot } from './sceneDebug.js'
import { initCollisionDebugPreference, isCollisionDebugEnabled, refreshCollisionDebugOverlay } from './collisionDebug.js'

export default class GameState {
  entdefs = new Map()
  hackHasLoadedInitialGrid = false
  gridScene = null
  grid = null
  pendingSpawns = []
  inventory = []
  itemDefinitions = new Map()
  recipeDefinitions = []

  hackHasLoadedLogo = false
  localPlayerEntityId = null
  collisionDebugEnabled = false

  constructor () {
    initCollisionDebugPreference(this)
  }

  addMessage (m) {}

  onNewGrid (g) {
    const current = this.gridScene
    // First grid (no scene yet) or mid-scroll: cut straight in.
    if (this.warpFading || (current != null && current.sys?.isActive?.() && !current.isScrollTransitioning)) {
      this.beginWarpFade(g)
      return
    }
    // Keep localPlayerEntityId across teleports; playerJoined is only sent once on connect.
    this.applyGrid(g, { resetLocalPlayer: false })
  }

  // Warps (teleport tiles, world loads) fade the old map to black, swap, then
  // fade the new map in. While fading, gridScene is null so the new map's
  // spawns queue in pendingSpawns (the same path as the first connect) instead
  // of landing in the old scene. A grid arriving mid-fade replaces the target.
  beginWarpFade (g) {
    this.pendingWarpGrid = g
    if (this.warpFading) {
      return
    }
    this.warpFading = true
    const old = this.gridScene
    this.gridScene = null
    old.fadeOutForWarp(() => {
      // The fade-complete event fires inside Phaser's scene update; swapping
      // scenes there is deferred by the scene manager and leaves the old scene
      // half torn down. Swap on the next task instead, outside the game step.
      setTimeout(() => {
        const next = this.pendingWarpGrid
        this.pendingWarpGrid = null
        this.warpFading = false
        this.fadeInNextGrid = true
        this.applyGrid(next, { resetLocalPlayer: false })
      }, 0)
    })
  }

  onGridTransition (g) {
    if (this.gridScene?.beginScrollTransition?.(g, g.transition)) {
      this.grid = g
      return
    }

    this.applyGrid(g, { resetLocalPlayer: false })
  }

  applyGrid (g, { resetLocalPlayer }) {
    const manager = window.phaser.scene

    logGreyvar('applyGrid()', {
      tileCount: g?.tiles?.length,
      entityCount: g?.entities?.length,
      gridId: g?.gridId,
      resetLocalPlayer
    })
    sceneSnapshot('applyGrid (before)')

    if (manager.getScene('grid')) {
      logGreyvar('applyGrid: removing existing grid scene')
      manager.stop('grid')
      manager.remove('grid')
    }

    const gs = new GridScene(g)
    const added = manager.add('grid', gs, true)

    logGreyvar('applyGrid: manager.add(grid)', {
      addedImmediately: added != null,
      isProcessing: manager.isProcessing,
      pendingCount: manager._pending?.length ?? 'n/a'
    })

    this.grid = g
    this.gridScene = gs
    this.hackHasLoadedInitialGrid = true

    if (resetLocalPlayer) {
      this.localPlayerEntityId = null
    }

    for (const ent of this.pendingSpawns) {
      this.gridScene.onEntitySpawn(ent)
    }
    this.pendingSpawns = []

    this.showGame('applyGrid')
    window.resizeGame()
    sceneSnapshot('applyGrid (after showGame scheduled)')

    if (isCollisionDebugEnabled()) {
      refreshCollisionDebugOverlay()
    }
  }

  showGame (source = 'unknown') {
    logGreyvar(`showGame() from ${source}`, {
      alreadyPending: this._showGamePending,
      stack: new Error().stack?.split('\n').slice(1, 5)
    })

    if (this._showGamePending) {
      return
    }

    this._showGamePending = true
    requestAnimationFrame(() => {
      this.finishShowGame(source, 0)
    })
  }

  hookGridLoadEvents (grid) {
    if (this._gridLoadEventsHooked) {
      return
    }

    this._gridLoadEventsHooked = true

    grid.load.on('loaderror', (file) => {
      logGreyvar('grid loaderror', {
        key: file?.key,
        url: file?.url,
        src: file?.src
      })
    })

    grid.load.on('complete', () => {
      logGreyvar('grid loader complete')
    })

    grid.events.once('create', () => {
      logGreyvar('grid scene create event fired')
      sceneSnapshot('grid create')
      if (this._showGamePending) {
        this.finishShowGame('grid-create-event', 0)
      }
    })
  }

  finishShowGame (source = 'unknown', attempt = 0) {
    const manager = window.phaser.scene
    const maxAttempts = 600

    if (attempt === 0) {
      logGreyvar(`finishShowGame() from ${source}`)
      sceneSnapshot('finishShowGame (before)')
    }

    if (!manager.isActive('hud')) {
      logGreyvar('finishShowGame: starting hud')
      manager.run('hud')
    }

    const grid = manager.getScene('grid')
    if (!grid) {
      if (attempt < maxAttempts) {
        requestAnimationFrame(() => this.finishShowGame(source, attempt + 1))
      } else {
        this._showGamePending = false
        logGreyvar('finishShowGame: gave up waiting for grid scene registration')
      }
      return
    }

    this.hookGridLoadEvents(grid)

    if (!manager.isActive('grid')) {
      if (attempt === 0 || attempt % 60 === 0) {
        logGreyvar('finishShowGame: waiting for grid to become active', {
          attempt,
          status: grid.sys.settings.status,
          loading: grid.sys.load?.isLoading?.(),
          list: grid.sys.load?.list?.length
        })
      }

      if (attempt < maxAttempts) {
        requestAnimationFrame(() => this.finishShowGame(source, attempt + 1))
      } else {
        this._showGamePending = false
        logGreyvar('finishShowGame: gave up waiting for grid to become active')
        sceneSnapshot('finishShowGame stuck waiting for grid')
      }
      return
    }

    this._showGamePending = false

    manager.bringToTop('grid')
    manager.bringToTop('hud')
    // world-load / teleports rebuild the grid scene, often from the console.
    // Keep an open menu or console above it; otherwise it is drawn under the
    // map, invisible but still swallowing keyboard input.
    for (const key of ['menu', 'console']) {
      if (manager.isActive(key)) {
        manager.bringToTop(key)
      }
    }

    const splash = manager.getScene('splash')
    if (splash) {
      splash.finished = true
      logGreyvar('finishShowGame: removing splash after grid is running', {
        splashStatus: splash.sys.settings.status,
        splashVisible: splash.sys.settings.visible
      })

      if (splash.sys.settings.status !== 8) {
        manager.stop('splash')
      }
      manager.remove('splash')
    }

    sceneSnapshot('finishShowGame (after grid ready)')
  }

  // The server restarted or the socket dropped: the next session assigns new
  // entity ids and resends the player, grid and spawns, so forget the old ones.
  onReconnected () {
    this.localPlayerEntityId = null
    this.pendingSpawns = []
  }

  onPlayerJoined (plj) {
    console.log('player joined:', plj.username)

    // playerJoined is sent when we connect; other players' joins must not
    // replace our local entity id or movement/camera attach to the wrong player.
    if (plj.entityId != null && this.localPlayerEntityId == null) {
      this.localPlayerEntityId = entityId(plj.entityId)
      // Bring back the look chosen in an earlier session (this browser).
      const restore = restoreRequest(loadSavedAppearance())
      if (restore != null) {
        window.serverConnection?.sendAppearanceRequest(restore)
      }
    }
  }

  onEntitySpawn (ent) {
    const normalized = normalizeEntitySpawn(ent)

    if (normalized.entityId == null) {
      return
    }

    if (this.gridScene?.isScrollTransitioning) {
      this.gridScene.queueTransitionSpawn(normalized)
      return
    }

    if (this.gridScene == null) {
      this.pendingSpawns.push(normalized)
      return
    }

    this.gridScene.onEntitySpawn(normalized)
  }

  onEntityPosition (entpos) {
    if (this.gridScene == null) {
      return
    }

    // Normalize first: proto3 omits zero-valued fields, so an entry at x=0 or
    // y=0 (right/down edge transitions) would otherwise arrive as undefined.
    const normalized = normalizeEntityPosition(entpos)

    if (this.gridScene.handleTransitionEntityPosition?.(normalized)) {
      return
    }

    this.gridScene.onEntityPosition(normalized)
  }

  onEntityDespawn (rawId) {
    const id = entityId(rawId)
    if (id == null || this.gridScene == null) {
      return
    }

    if (this.gridScene.isScrollTransitioning) {
      // Keep old-grid entities on screen until the scroll finishes.
      if (id !== this.localPlayerEntityId) {
        this.gridScene.queueTransitionDespawn(id)
      }
      return
    }

    this.gridScene.onEntityDespawn(id)
  }

  onEntdef (entdef) {
    if (entdef?.name == null) {
      return
    }

    const existing = this.entdefs.get(entdef.name)
    // Prefer keeping a definition that already has frame data if the update is empty.
    if (existing?.states?.some((s) => s?.frames?.length > 0) &&
        !entdef.states?.some((s) => s?.frames?.length > 0)) {
      return
    }

    this.entdefs.set(entdef.name, entdef)
  }

  onInventoryUpdate (update) {
    this.inventory = Array.isArray(update?.items) ? update.items.map((item) => ({
      definition: item.definition ?? '',
      count: item.count ?? 1
    })) : []
    logGreyvar('inventory update', { items: this.inventory })
  }

  onItemDefinition (itemdef) {
    if (itemdef?.id == null) {
      return
    }

    this.itemDefinitions.set(itemdef.id, {
      id: itemdef.id,
      title: itemdef.title ?? itemdef.id,
      stackMax: itemdef.stackMax ?? 99,
      category: itemdef.category ?? '',
      heldSprite: itemdef.heldSprite ?? '',
      heldGrip: [itemdef.heldGripX ?? 0, itemdef.heldGripY ?? 0]
    })
  }

  onRecipeDefinition (recipe) {
    if (recipe?.id == null) {
      return
    }

    const existing = this.recipeDefinitions.find((r) => r.id === recipe.id)
    if (existing != null) {
      return
    }

    this.recipeDefinitions.push({
      id: recipe.id,
      title: recipe.title ?? recipe.id,
      inputs: (recipe.inputs ?? []).map((input) => ({
        item: input.item ?? '',
        count: input.count ?? 1
      })),
      outputs: (recipe.outputs ?? []).map((output) => ({
        item: output.item ?? '',
        count: output.count ?? 1
      }))
    })
  }
}
