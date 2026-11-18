import Phaser from 'phaser'
import { initMovementInput, pollMovementInput, isMovementKeyDown } from '../MovementInput.js'
import { pollInteractInput } from '../InteractInput.js'
import LayeredSprite from '../sprites/LayeredSprite.js'
import { preloadRig, getRig } from '../sprites/PlayerRig.js'
import { TILE_SIZE, ENTITY_SIZE, isCollisionDebugEnabled, refreshCollisionDebugOverlay, tileCollisionRectsWorld, entityCollisionRectsWorld, isLegacyTileBlocked } from '../collisionDebug.js'

const MOVE_DURATION_MS = 67
// Constant-speed screen scroll (1.5x the ALttP-ish ~1s per screen).
const SCROLL_TRANSITION_MS_PER_SCREEN = 667
const SCROLL_TRANSITION_MIN_MS = 267

export default class GridScene extends Phaser.Scene {
  grid = null
  preloaded = false

  entities = {}

  isScrollTransitioning = false
  pendingTransitionSpawns = []
  pendingTransitionDespawns = []
  activeTransition = null

  constructor (grid) {
    super({ key: 'grid' })
    this.grid = grid
  }


  preload() {
    this.load.setBaseURL('/res/')
    this.load.image('construct.png', 'img/textures/tiles/construct.png')

    for (const entdef of window.gameState.entdefs.entries()) {
      this.preloadEntdefSprite(entdef[1])
    }

    // Players render from the layered rig (res/sprites/player); the entdef's
    // composite sheet above is only a fallback if the rig fails to load.
    preloadRig(this)

    this.load.atlas('water', 'img/textures/fluids/water.png', 'atlas/water.json')

    this.layerGrid = this.add.layer()
    this.layerEntities = this.add.layer()

    this.preloadTilesets()
    this.preloadAllGridTextures()
    this.preloaded = true
  }

  preloadEntdefSprite(entdef) {
    if (entdef?.texture == null || entdef.texture === '') {
      return
    }

    // Key by texture file so pot/key/chest sharing general-entities.png use one
    // spritesheet. Loading the same URL under many entity names breaks frames.
    const textureKey = entdef.texture
    if (this.textures.exists(textureKey)) {
      return
    }

    const frameSize = entdef.name === 'player' ? 16 : 15
    this.load.spritesheet(textureKey, 'img/textures/entities/' + entdef.texture, {
      frameWidth: frameSize,
      frameHeight: frameSize
    })
  }

  create() {
    console.log('[Greyvar] GridScene.create()', {
      tileCount: this.grid?.allTiles?.()?.length,
      entdefCount: window.gameState.entdefs.size
    })

    this.anims.create({
      key: 'water',
      frames: this.anims.generateFrameNames('water', { prefix: '', end: 2}),
      repeat: -1,
      duration: 1500,
    })

    this.pointer = this.input.activePointer

    this.renderGrid()
    this.renderEntities()
    this.configureCamera()
    this.entitiesReady = true

    initMovementInput(this)
    this.input.keyboard.enabled = true

    if (isCollisionDebugEnabled()) {
      refreshCollisionDebugOverlay()
    }
  }

  canRenderEntity () {
    return !this.load.isLoading()
  }

  spriteKeyForEntity (ent) {
    const entdef = window.gameState.entdefs.get(ent.definition)
    return entdef?.texture || ent.definition
  }

  initialStateForEntity (ent) {
    if (ent.initialState) {
      return ent.initialState
    }

    const entdef = window.gameState.entdefs.get(ent.definition)
    return entdef?.states?.[0]?.name ?? 'idle'
  }

  animKeyForEntity (definition, stateName) {
    return `${definition}:${stateName}`
  }

  stateFramesForEntity (ent, entdef) {
    if (entdef?.states == null) {
      return null
    }

    const stateName = this.initialStateForEntity(ent)
    return entdef.states.find((s) => s?.name === stateName)
      || entdef.states.find((s) => s?.name === ent.state)
      || entdef.states[0]
      || null
  }

  configureCamera () {
    const tileSize = 16
    const worldW = this.grid.colCount * tileSize
    const worldH = this.grid.rowCount * tileSize
    const cam = this.cameras.main

    // Even integer viewport avoids half-pixel camera origins under RESIZE.
    const viewW = Math.max(2, Math.floor(this.scale.width / 2) * 2)
    const viewH = Math.max(2, Math.floor(this.scale.height / 2) * 2)
    cam.setBounds(0, 0, worldW, worldH)
    cam.setSize(viewW, viewH)
    cam.setZoom(6)
    cam.roundPixels = true
    this.patchCameraScrollSnap(cam)
    this.snapCameraScroll()
  }

  // Camera follow runs inside Camera.preRender (after Scene.update). Snapping in
  // update() is overwritten before draw — patch so we snap after follow, before render.
  patchCameraScrollSnap (cam) {
    if (cam == null || cam._greyvarScrollSnapPatched) {
      return
    }

    const originalPreRender = cam.preRender.bind(cam)
    cam.preRender = () => {
      originalPreRender()
      this.snapCameraScroll()
    }
    cam._greyvarScrollSnapPatched = true
  }

  snapCameraScroll () {
    const cam = this.cameras.main
    if (cam == null) {
      return
    }
    const z = cam.zoom || 1
    cam.scrollX = Math.round(cam.scrollX * z) / z
    cam.scrollY = Math.round(cam.scrollY * z) / z
  }

  // Cover sub-pixel atlas sampling gaps under camera zoom (1 display pixel).
  applyTileSeamBleed (img, tileSize) {
    const zoom = Math.max(1, this.cameras.main?.zoom ?? 1)
    const size = tileSize + (1 / zoom)
    img.setDisplaySize(size, size)
  }

  mnuClicked() {
    console.log("mnu!")
  }

  update() {
    if (this.isScrollTransitioning) {
      return
    }

    pollMovementInput()
    pollInteractInput()
    this.updateLocalPlayerIdleAnimation()
    this.updateGamepad()

    for (const id of Object.keys(this.entities)) {
      if (this.entities[id].img == null) {
        this.renderEntity(this.entities[id])
      }
    }

    if (isCollisionDebugEnabled() && this.collisionDebugPlayerOutline != null) {
      const localId = window.gameState?.localPlayerEntityId
      const ent = localId != null ? this.entities[localId] : null
      if (ent?.img != null) {
        this.collisionDebugPlayerOutline.setPosition(ent.img.x, ent.img.y)
      }
    }
  }

  updateInputPointer(){
    let p = this.input.activePointer

    if (p.isDown) {
      if (!this.scale.fullscreen.active) {
//        this.scale.startFullscreen()
      }

      window.serverConnection.sendMoveRequest({x: 0, y: 1});
    }
  }

  updateGamepad() {
    if (this.input.gamepad.total === 0) {
      return
    }

    var pad = this.input.gamepad.getPad(0)

    if (pad.axes.length) {
      var x = pad.axes[6].getValue()
      var y = pad.axes[7].getValue()

      console.log(x, y)

      window.serverConnection.sendMoveRequest({x: x, y: y})
    }
  }

  renderEntities() {
    for (const ent of Object.keys(this.entities)) {
      this.renderEntity(this.entities[ent])
    }
  }

  renderEntity(ent) {
    if (ent.img != null) {
      return
    }

    if (!this.canRenderEntity()) {
      return
    }

    const entdef = window.gameState.entdefs.get(ent.definition)
    if (entdef == null) {
      return
    }

    if (ent.definition === 'player') {
      const rig = getRig(this)
      if (rig != null) {
        this.renderLayeredEntity(ent, rig)
        return
      }
    }

    const spriteKey = this.spriteKeyForEntity(ent)
    if (!spriteKey || !this.textures.exists(spriteKey)) {
      this.preloadEntdefSprite(entdef)
      if (this.load.isLoading()) {
        this.load.once('complete', () => this.renderEntity(ent))
        this.load.start()
      }
      return
    }

    const state = this.stateFramesForEntity(ent, entdef)
    const frame = state?.frames?.[0]
    const pos = this.entityRenderPosition(ent)
    const img = frame != null
      ? this.add.sprite(pos.x, pos.y, spriteKey, frame)
      : this.add.sprite(pos.x, pos.y, spriteKey)

    this.createEntityAnimations(img, ent.definition)

    img.setOrigin(0)
    const stateName = state?.name ?? this.initialStateForEntity(ent)
    const animKey = this.animKeyForEntity(ent.definition, stateName)
    if (img.anims?.exists(animKey) || this.anims?.exists(animKey)) {
      img.play(animKey)
    } else if (frame != null) {
      img.setFrame(frame)
    }

    ent.img = img

    this.layerEntities.add(ent.img)
    this.ensureCameraFollow(ent)
  }

  renderLayeredEntity (ent, rig) {
    const pos = this.entityRenderPosition(ent)
    const state = ent.state || this.initialStateForEntity(ent)
    const img = new LayeredSprite(this, pos.x, pos.y, rig, ent.appearance, state,
      (itemId) => window.gameState?.itemDefinitions?.get(itemId))
    ent.img = img
    this.layerEntities.add(img)
    this.ensureCameraFollow(ent)
  }

  ensureCameraFollow (ent) {
    if (ent.definition !== 'player' || ent.img == null) {
      return
    }

    const localId = window.gameState.localPlayerEntityId
    if (localId == null || ent.entityId !== localId) {
      return
    }

    const cam = this.cameras.main
    if (cam._follow === ent.img) {
      return
    }

    cam.centerOn(ent.img.x, ent.img.y)
    this.snapCameraScroll()
    cam.startFollow(ent.img, true, 0.15, 0.15)
    this.playerFollowTarget = ent.img
  }

  createEntityAnimations(img, definition) {
    const entdef = window.gameState.entdefs.get(definition)
    if (entdef == null) {
      return
    }

    const textureKey = entdef.texture || definition

    for (const state of entdef.states) {
      if (state?.name == null || !Array.isArray(state.frames) || state.frames.length === 0) {
        continue
      }

      const animKey = this.animKeyForEntity(definition, state.name)
      if (img.anims.exists(animKey)) {
        continue
      }

      const isWalk = String(state.name).startsWith('walk')
      const defaultFrameRate = isWalk ? 10 : 2
      const defaultHoldMs = Math.round(1000 / defaultFrameRate)
      const holds = Array.isArray(state.holds) ? state.holds : []
      // Tiled walk clips often ship duration=0; treat that as "use frameRate".
      const hasHolds = holds.some((hold) => hold > 0)

      const frames = state.frames.map((frame, index) => {
        const entry = { key: textureKey, frame }
        const hold = holds[index]
        if (hold > 0) {
          entry.duration = hold
        } else if (hasHolds) {
          entry.duration = defaultHoldMs
        }
        return entry
      })

      const anim = {
        key: animKey,
        frames,
        // Walk must loop: moves arrive every ~67ms and would otherwise restart
        // a one-shot clip on frame 0 forever (looks static).
        repeat: -1
      }
      if (!hasHolds) {
        anim.frameRate = defaultFrameRate
      }

      img.anims.create(anim)
    }
  }

  renderGrid() {
    if (this.grid.layers?.length > 0) {
      for (const layerDef of this.grid.layers) {
        const phaserLayer = this.add.layer()
        this.layerGrid.add(phaserLayer)
        this.renderTiles(layerDef.tiles, { offsetX: 0, offsetY: 0, layer: phaserLayer })
      }
      return
    }

    this.renderGridIntoLayer(this.grid, 0, 0, this.layerGrid)
  }

  renderGridIntoLayer (grid, offsetX, offsetY, layer) {
    this.renderTiles(grid.allTiles(), { offsetX, offsetY, layer })
  }

  renderTiles (tiles, opts = {}) {
    for (const tile of tiles) {
      this.renderTile(tile, opts)
    }
  }

  renderTile(tile, opts = {}) {
    const tileSize = 16
    const offsetX = opts.offsetX ?? 0
    const offsetY = opts.offsetY ?? 0
    const layer = opts.layer ?? this.layerGrid

    // Integer top-left cell origin. Keep vertex rounding off under camera zoom:
    // 'safe' rounding snaps each tile independently and opens 1px seams.
    const cellX = Math.round(offsetX + tile.col * tileSize)
    const cellY = Math.round(offsetY + tile.row * tileSize)
    const rotated = ((tile.textureRotation ?? 0) % 360) !== 0
    const flipped = !!(tile.textureHorizontalFlip || tile.textureVerticalFlip)
    const useCenter = rotated || flipped

    if (tile.textureName == 'water.png') {
      const water = this.add.sprite(cellX, cellY, 'water')
        .setOrigin(0)
        .setVertexRoundMode('off')
        .play('water')
      this.applyTileSeamBleed(water, tileSize)
      layer.add(water)
      return
    }

    let img
    if (tile.atlasKey && tile.frameIndex != null && tile.frameIndex >= 0) {
      if (useCenter) {
        img = this.add.image(cellX + tileSize / 2, cellY + tileSize / 2, tile.atlasKey, tile.frameIndex)
      } else {
        img = this.add.image(cellX, cellY, tile.atlasKey, tile.frameIndex).setOrigin(0)
      }
    } else if (useCenter) {
      img = this.add.image(cellX + tileSize / 2, cellY + tileSize / 2, tile.textureName)
    } else {
      img = this.add.image(cellX, cellY, tile.textureName).setOrigin(0)
    }

    img.setVertexRoundMode('off')
    this.applyTileSeamBleed(img, tileSize)
    img.angle = tile.textureRotation
    img.flipX = tile.textureHorizontalFlip
    img.flipY = tile.textureVerticalFlip
    layer.add(img)
  }

  preloadTilesets() {
    this.queueGridTextureLoads(this.grid)
  }

  preloadAllGridTextures() {
    this.queueGridTextureLoads(this.grid)
  }

  queueGridTextureLoads (grid) {
    for (const tileset of grid.tilesets ?? []) {
      if (!tileset.key || !tileset.image || this.textures.exists(tileset.key)) {
        continue
      }

      this.load.spritesheet(tileset.key, tileset.image, {
        frameWidth: tileset.tileWidth ?? 16,
        frameHeight: tileset.tileHeight ?? 16
      })
    }

    for (const tile of grid.allLayerTiles?.() ?? grid.allTiles()) {
      if (tile.atlasKey) {
        continue
      }

      if (tile.textureName?.includes('#')) {
        continue
      }

      if (this.textures.exists(tile.textureName)) {
        continue
      }

      this.load.image(tile.textureName, 'img/textures/tiles/' + tile.textureName)
    }
  }

  // ALttP-style screen scroll: the camera pans at a constant speed from where it
  // is to exactly where it will rest on the new grid (about one screen), while
  // the local player keeps walking a short way onto the new grid. Linear motion
  // with no overshoot or end snap keeps it readable and comfortable.
  beginScrollTransition (newGrid, transition) {
    const deltaX = transition?.scrollDeltaX ?? 0
    const deltaY = transition?.scrollDeltaY ?? 0

    if (!transition || (deltaX === 0 && deltaY === 0)) {
      return false
    }

    if (this.isScrollTransitioning) {
      return false
    }

    this.isScrollTransitioning = true
    this.pendingTransitionSpawns = []
    this.pendingTransitionDespawns = []

    const localEnt = this.localPlayerEntity()
    if (localEnt?.moveTween != null) {
      localEnt.moveTween.stop()
      localEnt.moveTween = null
    }

    this.activeTransition = {
      newGrid,
      deltaX,
      deltaY,
      // Old-grid coordinates. Predicted from the server's edge-entry rules and
      // replaced by the authoritative position when it arrives.
      playerTarget: this.predictTransitionEntry(localEnt, newGrid, deltaX, deltaY),
      playerStart: null
    }

    this.cameras.main.stopFollow()

    this.queueGridTextureLoads(newGrid)

    if (this.load.isLoading()) {
      this.load.once('complete', () => this.runScrollTransition())
      this.load.start()
      return true
    }

    this.runScrollTransition()
    return true
  }

  localPlayerEntity () {
    const localId = window.gameState?.localPlayerEntityId
    return localId != null ? this.entities[localId] ?? null : null
  }

  // Mirrors server tryEdgeTransition: the player enters at the facing edge of the
  // new grid, keeping their position on the other axis.
  predictTransitionEntry (localEnt, newGrid, deltaX, deltaY) {
    if (localEnt == null) {
      return null
    }

    const newW = newGrid.colCount * TILE_SIZE
    const newH = newGrid.rowCount * TILE_SIZE
    let localX = localEnt.x ?? localEnt.img?.x ?? 0
    let localY = localEnt.y ?? localEnt.img?.y ?? 0

    if (deltaX > 0) {
      localX = 0
    } else if (deltaX < 0) {
      localX = newW - ENTITY_SIZE
    }

    if (deltaY > 0) {
      localY = 0
    } else if (deltaY < 0) {
      localY = newH - ENTITY_SIZE
    }

    return { x: localX + deltaX, y: localY + deltaY }
  }

  runScrollTransition () {
    const t = this.activeTransition
    const { newGrid, deltaX, deltaY } = t
    const currentW = this.grid.colCount * TILE_SIZE
    const currentH = this.grid.rowCount * TILE_SIZE
    const newW = newGrid.colCount * TILE_SIZE
    const newH = newGrid.rowCount * TILE_SIZE
    const cam = this.cameras.main
    const zoom = cam.zoom || 1

    cam.stopFollow()

    this.incomingLayer = this.add.layer()
    if (newGrid.layers?.length > 0) {
      for (const layerDef of newGrid.layers) {
        this.renderTiles(layerDef.tiles, {
          offsetX: deltaX,
          offsetY: deltaY,
          layer: this.incomingLayer
        })
      }
    } else {
      this.renderGridIntoLayer(newGrid, deltaX, deltaY, this.incomingLayer)
    }
    // Draw the incoming grid beneath entities so the player walks over it.
    this.children.moveBelow(this.incomingLayer, this.layerEntities)

    this.snapCameraScroll()
    const startScrollX = cam.scrollX
    const startScrollY = cam.scrollY

    // Where the camera will rest once the new grid is live: the same
    // centre-on-player + clamp-to-bounds that finishScrollTransition applies,
    // expressed in old-grid coordinates. Ending exactly here avoids a snap.
    const localImg = this.localPlayerEntity()?.img ?? null
    const target = t.playerTarget ?? (localImg != null ? { x: localImg.x, y: localImg.y } : null)
    cam.setBounds(deltaX, deltaY, newW, newH)
    if (target != null) {
      cam.centerOn(target.x, target.y)
    } else {
      cam.centerToBounds()
    }
    const endScrollX = Math.round(cam.scrollX * zoom) / zoom
    const endScrollY = Math.round(cam.scrollY * zoom) / zoom

    const minX = Math.min(0, deltaX)
    const minY = Math.min(0, deltaY)
    const maxX = Math.max(currentW, deltaX + newW)
    const maxY = Math.max(currentH, deltaY + newH)
    cam.setBounds(minX, minY, maxX - minX, maxY - minY)
    cam.scrollX = startScrollX
    cam.scrollY = startScrollY

    if (localImg != null) {
      t.playerStart = { x: localImg.x, y: localImg.y }
      this.playWalkAnimation(localImg, 0, 0, Math.sign(deltaX), Math.sign(deltaY))
      // Keep walking for the whole scroll (the walk otherwise times out to idle).
      t.walkState = localImg.state
    }

    // Constant speed, scaled by how far the view actually travels relative to
    // its own size, so every screen edge takes the same time to cross.
    const viewW = cam.width / zoom
    const viewH = cam.height / zoom
    const screens = Math.max(
      Math.abs(endScrollX - startScrollX) / viewW,
      Math.abs(endScrollY - startScrollY) / viewH
    )
    const duration = Math.max(SCROLL_TRANSITION_MIN_MS, screens * SCROLL_TRANSITION_MS_PER_SCREEN)
    const progress = { p: 0 }

    this.tweens.add({
      targets: progress,
      p: 1,
      duration,
      ease: 'Linear',
      onUpdate: () => {
        const p = progress.p
        const x = startScrollX + (endScrollX - startScrollX) * p
        const y = startScrollY + (endScrollY - startScrollY) * p
        cam.scrollX = Math.round(x * zoom) / zoom
        cam.scrollY = Math.round(y * zoom) / zoom
        this.stepTransitionPlayer(p)
      },
      onComplete: () => {
        cam.scrollX = endScrollX
        cam.scrollY = endScrollY
        this.stepTransitionPlayer(1)
        this.finishScrollTransition()
      }
    })
  }

  // The player walks from where they stood to their entry point on the new grid
  // over the course of the scroll (typically about one tile).
  stepTransitionPlayer (p) {
    const t = this.activeTransition
    const img = this.localPlayerEntity()?.img
    if (t == null || img == null || t.playerStart == null || t.playerTarget == null) {
      return
    }

    img.x = Math.round(t.playerStart.x + (t.playerTarget.x - t.playerStart.x) * p)
    img.y = Math.round(t.playerStart.y + (t.playerTarget.y - t.playerStart.y) * p)
    if (t.walkState != null) {
      img.playState?.(t.walkState)
    }
  }

  finishScrollTransition () {
    const { newGrid } = this.activeTransition

    if (this.incomingLayer != null) {
      this.incomingLayer.destroy(true)
      this.incomingLayer = null
    }

    this.layerGrid.removeAll(true)

    // Old-grid entities stay visible while they scroll off screen; drop them
    // now, before any new-grid spawns that might reuse an id.
    for (const id of this.pendingTransitionDespawns) {
      this.onEntityDespawn(id)
    }
    this.pendingTransitionDespawns = []

    for (const ent of Object.values(this.entities)) {
      delete ent.transitionIncoming
      if (ent.img != null) {
        ent.img.x = ent.x
        ent.img.y = ent.y
      }
    }

    this.grid = newGrid
    window.gameState.grid = newGrid
    this.renderGrid()
    this.configureCamera()

    this.isScrollTransitioning = false
    this.activeTransition = null

    for (const ent of this.pendingTransitionSpawns) {
      this.onEntitySpawn(ent)
    }
    this.pendingTransitionSpawns = []

    const localEnt = this.localPlayerEntity()
    if (localEnt?.img != null) {
      const cam = this.cameras.main
      cam.centerOn(localEnt.img.x, localEnt.img.y)
      this.snapCameraScroll()
      this.ensureCameraFollow(localEnt)
    }

    if (isCollisionDebugEnabled()) {
      refreshCollisionDebugOverlay()
    }
  }

  // Where an entity's sprite sits right now. New-grid entities spawned during a
  // scroll are drawn offset into the old grid's frame so they slide in with
  // their map.
  entityRenderPosition (ent) {
    const t = this.activeTransition
    if (t != null && ent.transitionIncoming) {
      return { x: (ent.x ?? 0) + t.deltaX, y: (ent.y ?? 0) + t.deltaY }
    }
    return { x: ent.x ?? 0, y: ent.y ?? 0 }
  }

  queueTransitionSpawn (ent) {
    // Ids already on screen (the local player, or an old-grid entity still
    // scrolling out) are reconciled once the transition completes.
    if (this.entities[ent.entityId] != null) {
      this.pendingTransitionSpawns.push(ent)
      return
    }

    if (ent.state == null) {
      ent.state = this.initialStateForEntity(ent)
    }
    ent.transitionIncoming = true
    this.entities[ent.entityId] = ent

    if (this.canRenderEntity()) {
      this.renderEntity(ent)
    } else {
      this.load.once('complete', () => this.renderEntity(ent))
    }
  }

  queueTransitionDespawn (entityId) {
    if (this.entities[entityId]?.transitionIncoming) {
      this.onEntityDespawn(entityId)
      return
    }
    this.pendingTransitionDespawns.push(entityId)
  }

  handleTransitionEntityPosition (entpos) {
    if (!this.isScrollTransitioning || this.activeTransition == null) {
      return false
    }

    const localId = window.gameState.localPlayerEntityId
    if (localId == null) {
      return false
    }

    if (entpos.entityId !== localId) {
      const other = this.entities[entpos.entityId]
      if (!other?.transitionIncoming) {
        return false
      }

      other.x = entpos.x
      other.y = entpos.y
      if (other.img != null) {
        const pos = this.entityRenderPosition(other)
        other.img.setPosition(pos.x, pos.y)
      }
      return true
    }

    const ent = this.entities[localId]
    if (ent?.img == null) {
      return false
    }

    // Server position is in new-grid coordinates; the walk-in target is in the
    // old grid's frame until the transition completes.
    ent.x = entpos.x
    ent.y = entpos.y
    this.activeTransition.playerTarget = {
      x: entpos.x + this.activeTransition.deltaX,
      y: entpos.y + this.activeTransition.deltaY
    }

    if (ent.moveTween != null) {
      ent.moveTween.stop()
      ent.moveTween = null
    }

    return true
  }

  onEntityAppearance ({ entityId, appearance }) {
    const ent = this.entities[entityId]
    if (ent == null) {
      return
    }
    ent.appearance = appearance
    ent.img?.setAppearance?.(appearance)
  }

  renderConnectionResponse(r) {
    this.text = 'Server: ' + r.serverVersion
  }

  onEntitySpawn (ent) {
    if (ent.state == null) {
      ent.state = this.initialStateForEntity(ent)
    }

    const existing = this.entities[ent.entityId]
    if (existing != null) {
      existing.definition = ent.definition ?? existing.definition
      existing.state = ent.state ?? existing.state
      if (ent.appearance != null) {
        existing.appearance = ent.appearance
        existing.img?.setAppearance?.(ent.appearance)
      }
      existing.x = ent.x ?? existing.x
      existing.y = ent.y ?? existing.y
      if (existing.img != null) {
        existing.img.x = existing.x
        existing.img.y = existing.y
        return
      }
      this.renderEntity(existing)
      return
    }

    this.entities[ent.entityId] = ent
    this.renderEntity(ent)
    if (isCollisionDebugEnabled()) {
      refreshCollisionDebugOverlay()
    }
  }

  onEntityDespawn (entityId) {
    const ent = this.entities[entityId]
    if (ent?.img != null) {
      ent.img.destroy()
    }

    delete this.entities[entityId]
  }

  playWalkAnimation (img, fromX, fromY, toX, toY) {
    const dx = toX - fromX
    const dy = toY - fromY

    let state = null
    if (dx < 0 && dy < 0) {
      state = 'walkUpLeft'
    } else if (dx > 0 && dy < 0) {
      state = 'walkUpRight'
    } else if (dx < 0 && dy > 0) {
      state = 'walkDownLeft'
    } else if (dx > 0 && dy > 0) {
      state = 'walkDownRight'
    } else if (dx > 0) {
      state = 'walkRight'
    } else if (dx < 0) {
      state = 'walkLeft'
    } else if (dy > 0) {
      state = 'walkDown'
    } else if (dy < 0) {
      state = 'walkUp'
    }

    if (state != null && img.playState != null) {
      img.playState(state)
      return
    }

    if (state != null) {
      const animKey = this.animKeyForEntity('player', state)
      // Don't restart mid-stride on every MoveRequest (~67ms) or the clip
      // never advances past the first frame.
      img.play(animKey, true)
    }
  }

  updateLocalPlayerIdleAnimation () {
    if (isMovementKeyDown()) {
      return
    }

    const localId = window.gameState?.localPlayerEntityId
    if (localId == null) {
      return
    }

    const ent = this.entities[localId]
    const img = ent?.img
    if (img == null || ent.definition !== 'player') {
      return
    }

    if (img.playState != null) {
      img.playState('idle')
      return
    }

    const idleKey = this.animKeyForEntity('player', 'idle')
    const animKey = img.anims?.currentAnim?.key
    if (animKey != null && animKey !== idleKey) {
      img.play(idleKey)
    }
  }

  moveEntityTo (ent, x, y) {
    const img = ent.img
    if (img == null) {
      return
    }

    if (img.x === x && img.y === y) {
      return
    }

    const fromX = img.x
    const fromY = img.y

    if (ent.definition === 'player') {
      this.playWalkAnimation(img, fromX, fromY, x, y)
    }

    if (ent.moveTween != null) {
      ent.moveTween.stop()
      ent.moveTween = null
    }

    ent.moveTween = this.tweens.add({
      targets: img,
      x,
      y,
      duration: MOVE_DURATION_MS,
      ease: 'Linear',
      onComplete: () => {
        ent.moveTween = null
      }
    })
  }

  onEntityPosition (entpos) {
    const ent = this.entities[entpos.entityId]

    if (ent == null) {
      return
    }

    ent.x = entpos.x
    ent.y = entpos.y

    if (ent.img == null) {
      this.renderEntity(ent)
    }

    if (ent.img == null) {
      return
    }

    this.ensureCameraFollow(ent)
    this.moveEntityTo(ent, entpos.x, entpos.y)

    const localId = window.gameState?.localPlayerEntityId
    if (isCollisionDebugEnabled() && localId != null && entpos.entityId === localId) {
      this.updateCollisionDebugPlayerOutline(entpos.x, entpos.y)
    }
  }

  onEntityChange(ec) {
    if (!String(ec.newState).startsWith('bump')) {
      console.log('entchange', ec)
    }

    let ent = this.entities[ec.entityId]

    if (ent == null) {
      console.log('cannot find ent to change', ec.entityId)
      return
    }

    if (ec.newState) {
      // Bumps are momentary events; keep the entity's real state for re-renders.
      if (!String(ec.newState).startsWith('bump')) {
        ent.state = ec.newState
      }
      const animKey = this.animKeyForEntity(ent.definition, ec.newState)
      if (ent.img?.playState != null) {
        ent.img.playState(ec.newState)
      } else if (ent.img?.anims?.exists(animKey)) {
        ent.img.play(animKey)
      } else {
        const entdef = window.gameState.entdefs.get(ent.definition)
        const state = entdef?.states?.find((s) => s?.name === ec.newState)
        const frame = state?.frames?.[0]
        if (ent.img != null && frame != null) {
          ent.img.setFrame(frame)
        }
      }
    }

    if (isCollisionDebugEnabled()) {
      refreshCollisionDebugOverlay()
    }
  }

  showCollisionDebugOverlay () {
    this.hideCollisionDebugOverlay()
    this.layerCollisionDebug = this.add.layer()
    this.layerCollisionDebug.setDepth(20)

    for (let row = 0; row < this.grid.rowCount; row++) {
      for (let col = 0; col < this.grid.colCount; col++) {
        const tile = this.grid.topTileAt(row, col)
        if (tile == null) {
          continue
        }

        if (tile.collision?.length > 0) {
          for (const rect of tileCollisionRectsWorld(tile)) {
            const box = this.add.rectangle(
              rect.x,
              rect.y,
              rect.w,
              rect.h,
              0xff3333,
              0.55
            ).setOrigin(0)
            this.layerCollisionDebug.add(box)
          }
          continue
        }

        if (tile.collisionFromTsx) {
          continue
        }

        const blocked = isLegacyTileBlocked(tile)
        const cell = this.add.rectangle(
          col * TILE_SIZE,
          row * TILE_SIZE,
          TILE_SIZE,
          TILE_SIZE,
          blocked ? 0xff3333 : 0x33ff66,
          blocked ? 0.35 : 0.18
        ).setOrigin(0)
        this.layerCollisionDebug.add(cell)
      }
    }

    for (const ent of Object.values(this.entities)) {
      const entdef = window.gameState?.entdefs?.get(ent.definition)
      for (const rect of entityCollisionRectsWorld(ent, entdef)) {
        const box = this.add.rectangle(
          rect.x,
          rect.y,
          rect.w,
          rect.h,
          0xff9933,
          0.45
        ).setOrigin(0)
        this.layerCollisionDebug.add(box)
      }
    }

    const localId = window.gameState?.localPlayerEntityId
    const localEnt = localId != null ? this.entities[localId] : null
    if (localEnt?.img != null) {
      this.collisionDebugPlayerOutline = this.add.rectangle(
        localEnt.img.x,
        localEnt.img.y,
        ENTITY_SIZE,
        ENTITY_SIZE
      ).setOrigin(0).setStrokeStyle(1, 0x3399ff, 1).setFillStyle(0x3399ff, 0.12)
      this.layerCollisionDebug.add(this.collisionDebugPlayerOutline)
    }

    if (this.layerEntities) {
      this.layerEntities.setDepth(30)
    }
  }

  hideCollisionDebugOverlay () {
    if (this.layerCollisionDebug != null) {
      this.layerCollisionDebug.destroy(true)
      this.layerCollisionDebug = null
    }
    this.collisionDebugPlayerOutline = null
  }

  updateCollisionDebugPlayerOutline (x, y) {
    if (this.collisionDebugPlayerOutline != null) {
      this.collisionDebugPlayerOutline.setPosition(x, y)
    }
  }
}
