import Phaser from 'phaser'
import { initMovementInput, pollMovementInput, isMovementKeyDown } from '../MovementInput.js'
import { TILE_SIZE, ENTITY_SIZE, isCollisionDebugEnabled, refreshCollisionDebugOverlay, tileCollisionRectsWorld, entityCollisionRectsWorld, isLegacyTileBlocked } from '../collisionDebug.js'

const MOVE_DURATION_MS = 67
const SCROLL_TRANSITION_MS = 380

export default class GridScene extends Phaser.Scene {
  grid = null
  preloaded = false

  entities = {}

  isScrollTransitioning = false
  pendingTransitionSpawns = []

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
    const img = frame != null
      ? this.add.sprite(ent.x ?? 0, ent.y ?? 0, spriteKey, frame)
      : this.add.sprite(ent.x ?? 0, ent.y ?? 0, spriteKey)

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
    this.activeTransition = {
      newGrid,
      deltaX,
      deltaY
    }

    this.queueGridTextureLoads(newGrid)

    if (this.load.isLoading()) {
      this.load.once('complete', () => this.runScrollTransition())
      this.load.start()
      return true
    }

    this.runScrollTransition()
    return true
  }

  runScrollTransition () {
    const { newGrid, deltaX, deltaY } = this.activeTransition
    const tileSize = 16
    const currentW = this.grid.colCount * tileSize
    const currentH = this.grid.rowCount * tileSize
    const newW = newGrid.colCount * tileSize
    const newH = newGrid.rowCount * tileSize
    const cam = this.cameras.main

    cam.stopFollow()

    const minX = Math.min(0, deltaX)
    const minY = Math.min(0, deltaY)
    const maxX = Math.max(currentW, deltaX + newW)
    const maxY = Math.max(currentH, deltaY + newH)
    cam.setBounds(minX, minY, maxX - minX, maxY - minY)

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

    this.snapCameraScroll()
    const startScrollX = cam.scrollX
    const startScrollY = cam.scrollY
    const endScrollX = startScrollX + deltaX
    const endScrollY = startScrollY + deltaY
    const zoom = cam.zoom || 1
    const progress = { t: 0 }

    this.tweens.add({
      targets: progress,
      t: 1,
      duration: SCROLL_TRANSITION_MS,
      ease: 'Cubic.easeInOut',
      onUpdate: () => {
        const x = startScrollX + (endScrollX - startScrollX) * progress.t
        const y = startScrollY + (endScrollY - startScrollY) * progress.t
        cam.scrollX = Math.round(x * zoom) / zoom
        cam.scrollY = Math.round(y * zoom) / zoom
      },
      onComplete: () => {
        cam.scrollX = Math.round(endScrollX * zoom) / zoom
        cam.scrollY = Math.round(endScrollY * zoom) / zoom
        this.finishScrollTransition()
      }
    })
  }

  finishScrollTransition () {
    const { newGrid, deltaX, deltaY } = this.activeTransition

    if (this.incomingLayer != null) {
      this.incomingLayer.destroy(true)
      this.incomingLayer = null
    }

    this.layerGrid.removeAll(true)

    for (const ent of Object.values(this.entities)) {
      if (ent.img != null) {
        ent.img.x -= deltaX
        ent.img.y -= deltaY
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

    const localId = window.gameState.localPlayerEntityId
    const localEnt = localId != null ? this.entities[localId] : null
    if (localEnt?.img != null) {
      this.ensureCameraFollow(localEnt)
    }

    if (isCollisionDebugEnabled()) {
      refreshCollisionDebugOverlay()
    }
  }

  queueTransitionSpawn (ent) {
    this.pendingTransitionSpawns.push(ent)
  }

  handleTransitionEntityPosition (entpos) {
    if (!this.isScrollTransitioning || this.activeTransition == null) {
      return false
    }

    const localId = window.gameState.localPlayerEntityId
    if (localId == null || entpos.entityId !== localId) {
      return false
    }

    const ent = this.entities[localId]
    if (ent?.img == null) {
      return false
    }

    const { deltaX, deltaY } = this.activeTransition
    ent.x = entpos.x
    ent.y = entpos.y

    if (ent.moveTween != null) {
      ent.moveTween.stop()
      ent.moveTween = null
    }

    ent.moveTween = this.tweens.add({
      targets: ent.img,
      x: entpos.x + deltaX,
      y: entpos.y + deltaY,
      duration: SCROLL_TRANSITION_MS,
      ease: 'Cubic.easeInOut',
      onComplete: () => {
        ent.moveTween = null
      }
    })

    return true
  }

  renderConnectionResponse(r) {
    this.text = 'Server: ' + r.serverVersion
  }

  onEntitySpawn (ent) {
    if (ent.state == null) {
      ent.state = this.initialStateForEntity(ent)
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
    console.log('entchange', ec)

    let ent = this.entities[ec.entityId]

    if (ent == null) {
      console.log('cannot find ent to change', ec.entityId)
      return
    }

    if (ec.newState) {
      ent.state = ec.newState
      const animKey = this.animKeyForEntity(ent.definition, ec.newState)
      if (ent.img?.anims?.exists(animKey)) {
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
