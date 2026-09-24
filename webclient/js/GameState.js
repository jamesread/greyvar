import GridScene from './scenes/GridScene.js'
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

  hackHasLoadedLogo = false
  localPlayerEntityId = null
  collisionDebugEnabled = false

  constructor () {
    initCollisionDebugPreference(this)
  }

  addMessage (m) {}

  onNewGrid (g) {
    this.applyGrid(g, { resetLocalPlayer: true })
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

  onPlayerJoined (plj) {
    console.log('player joined:', plj.username)

    if (plj.entityId != null) {
      this.localPlayerEntityId = entityId(plj.entityId)
    }
  }

  onEntitySpawn (ent) {
    const normalized = normalizeEntitySpawn(ent)

    if (normalized.entityId == null) {
      return
    }

    if (normalized.definition === 'player') {
      if (this.localPlayerEntityId == null || normalized.entityId > this.localPlayerEntityId) {
        this.localPlayerEntityId = normalized.entityId
      }
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

    if (this.gridScene.handleTransitionEntityPosition?.(entpos)) {
      return
    }

    const normalized = normalizeEntityPosition(entpos)
    this.gridScene.onEntityPosition(normalized)
  }

  onEntityDespawn (rawId) {
    const id = entityId(rawId)
    if (id == null || this.gridScene == null) {
      return
    }

    if (this.gridScene.isScrollTransitioning && id === this.localPlayerEntityId) {
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
}
