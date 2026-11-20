import { showDialog } from './dialog.js'
import { greyvarproto } from './greyvarproto.js'
import { buildGridFromMessage } from './gridFromMessage.js'
import { normalizeEntityStateChange } from './entityUtils.js'
import { logGreyvar } from './sceneDebug.js'
import { reconnectDelayMs, setDisconnectedBanner } from './reconnect.js'

export default class ServerConnection {
  isOk = false
  // Reconnect state: attempts since the last good connection, and whether we
  // have ever connected (a reconnect must reset session state on the client).
  reconnectAttempt = 0
  hasConnected = false
  closedByUser = false
  reconnectTimer = null

  constructor () {
    this.connect()
  }

  connect () {
    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const address = `${protocol}//${window.location.host}/api`

      console.log('address', address)

      this.sock = new WebSocket(address)
      this.sock.onclose = () => this.onClose()
      this.sock.onerror = (e) => console.error('ServerConnection error', e)
    } catch (e) {
      console.error(e)
    }

    this.sock.onopen = () => {
      this.isOk = true
      if (this.hasConnected) {
        // New server session: fresh entity ids, player and grid follow.
        window.gameState?.onReconnected?.()
      }
      this.hasConnected = true
      this.reconnectAttempt = 0
      setDisconnectedBanner(false)

      this.sendClientRequest({
        registrationRequest: {
          username: 'james'
        }
      })
    }

    this.sock.onmessage = (m) => this.onMessage(m)
  }

  sendClientRequest (request) {
    if (this.sock?.readyState !== WebSocket.OPEN) {
      return
    }

    // Send request JSON directly so new proto fields work before greyvarproto.js is regenerated.
    this.sock.send(JSON.stringify(request))
  }

  onMessage (m) {
    let receivedMessage

    try {
      receivedMessage = JSON.parse(m.data)
    } catch (e) {
      console.error('ServerConnection failed to parse message', e, m.data)
      return
    }

    if (receivedMessage.connectionResponse != null) {
      window.gameState.addMessage(receivedMessage.connectionResponse)
    }

    for (const entdef of receivedMessage.entityDefinitions ?? []) {
      window.gameState.onEntdef(entdef)
    }

    const newGrid = buildGridFromMessage(receivedMessage.grid)

    if (newGrid != null) {
      const previousGridId = window.gameState.grid?.gridId ?? null
      const previousWorldId = window.gameState.grid?.worldId ?? null
      const nextGridId = newGrid.gridId ?? receivedMessage.grid?.gridId ?? receivedMessage.grid?.title ?? null
      const nextWorldId = newGrid.worldId ?? receivedMessage.grid?.worldId ?? null
      const isTransition = previousGridId != null && (
        (nextGridId != null && previousGridId !== nextGridId) ||
        (nextWorldId != null && previousWorldId !== nextWorldId)
      )
      const scrollX = newGrid.transition?.scrollDeltaX ?? 0
      const scrollY = newGrid.transition?.scrollDeltaY ?? 0
      const hasScrollTransition = isTransition && newGrid.transition != null && (scrollX !== 0 || scrollY !== 0)

      logGreyvar('ServerConnection: grid message received', {
        tileCount: newGrid.tiles?.length,
        entityCount: newGrid.entities?.length,
        gridId: nextGridId,
        isTransition,
        hasScrollTransition,
        transition: newGrid.transition,
        hasGridField: receivedMessage.grid != null
      })

      // Adjacent-map walks use a scroll tween. Teleports / world loads are a hard cut
      // (onNewGrid) so input and the local player are fully reset.
      if (hasScrollTransition) {
        window.gameState.onGridTransition(newGrid)
      } else {
        window.gameState.onNewGrid(newGrid)
      }
    }

    for (const rawId of receivedMessage.entityDespawns ?? []) {
      window.gameState.onEntityDespawn(rawId)
    }

    if (receivedMessage.worldListResponse?.worlds != null) {
      this.onWorldListResponse(receivedMessage.worldListResponse)
    }

    if (receivedMessage.consoleMessage?.text != null) {
      window.gameConsole?.log(receivedMessage.consoleMessage.text)
    }

    if (receivedMessage.dialog != null) {
      showDialog(receivedMessage.dialog)
    }

    if (receivedMessage.hudMessage != null) {
      window.hudScene?.setHudMessage?.(receivedMessage.hudMessage)
    }

    if (receivedMessage.inventoryUpdate?.items != null) {
      window.gameState.onInventoryUpdate?.(receivedMessage.inventoryUpdate)
    }

    for (const itemdef of receivedMessage.itemDefinitions ?? []) {
      window.gameState.onItemDefinition?.(itemdef)
    }

    for (const recipe of receivedMessage.recipeDefinitions ?? []) {
      window.gameState.onRecipeDefinition?.(recipe)
    }

    if (receivedMessage.playerJoined != null) {
      window.gameState.onPlayerJoined(receivedMessage.playerJoined)
    }

    for (const ent of receivedMessage.entitySpawns ?? []) {
      window.gameState.onEntitySpawn(ent)
    }

    if (window.gameState.gridScene == null) {
      return
    }

    for (const entpos of receivedMessage.entityPositions ?? []) {
      window.gameState.onEntityPosition(entpos)
    }

    for (const entchange of receivedMessage.entityStateChanges ?? []) {
      window.gameState.gridScene.onEntityChange(normalizeEntityStateChange(entchange))
    }

    for (const change of receivedMessage.tileChanges ?? []) {
      window.gameState.gridScene.onTileChange(change)
    }

    for (const change of receivedMessage.entityAppearanceChanges ?? []) {
      window.gameState.gridScene.onEntityAppearance(normalizeEntityStateChange(change))
    }
  }

  sendInteractRequest (entityId, cell = null, facing = null) {
    const req = { entityId: Number(entityId ?? 0) }
    if (cell != null) {
      req.cell = { row: cell.row, col: cell.col }
    }
    if (facing) {
      req.facing = String(facing)
    }
    if (!req.entityId && req.cell == null) {
      return
    }
    this.sendClientRequest({ interactRequest: req })
  }

  sendDialogChoice (entityId, node, option) {
    this.sendClientRequest({ dialogChoiceRequest: { entityId: Number(entityId), node: String(node ?? ''), option: Number(option) } })
  }

  sendCraftRequest (recipeId) {
    if (recipeId == null || recipeId === '') {
      return
    }

    this.sendClientRequest({
      craftRequest: { recipeId: String(recipeId) }
    })
  }

  sendMoveRequest (vec) {
    const x = Math.trunc(vec.x ?? 0)
    const y = Math.trunc(vec.y ?? 0)

    if (x === 0 && y === 0) {
      return
    }

    this.sendClientRequest({
      moveRequest: greyvarproto.MoveRequest.create({ x, y })
    })
  }

  sendWorldListRequest () {
    this.sendClientRequest({
      worldListRequest: {}
    })
  }

  sendWorldLoadRequest (worldId) {
    this.sendClientRequest({
      worldLoadRequest: { worldId }
    })
  }

  /** @param {{level: number, seed: number, mechanics: string[], leave: boolean}} req */
  sendRiftRequest (req) {
    this.sendClientRequest({ riftRequest: req })
  }

  sendWorldReloadRequest (worldId) {
    const request = { worldReloadRequest: {} }
    if (worldId != null && worldId !== '') {
      request.worldReloadRequest.worldId = String(worldId)
    }
    this.sendClientRequest(request)
  }

  /** @param {{parts?: object, palettes?: object, heldItem?: string}} req */
  sendAppearanceRequest (req) {
    this.sendClientRequest({ appearanceRequest: req })
  }

  disconnect () {
    this.closedByUser = true
    clearTimeout(this.reconnectTimer)
    this.isOk = false
    this.sock?.close()
  }

  onClose () {
    this.isOk = false
    if (this.closedByUser) {
      return
    }
    // Server restarts (air rebuilds) close the socket; keep trying.
    const delay = reconnectDelayMs(this.reconnectAttempt++)
    console.warn(`ServerConnection closed; reconnecting in ${delay}ms`)
    if (this.hasConnected) {
      setDisconnectedBanner(true)
    }
    clearTimeout(this.reconnectTimer)
    this.reconnectTimer = setTimeout(() => this.connect(), delay)
  }

  onWorldListResponse (response) {
    const worlds = response.worlds ?? []
    if (worlds.length === 0) {
      window.gameConsole?.log('no worlds found')
      return
    }

    window.gameConsole?.log(`worlds (${worlds.length}):`)
    for (const world of worlds) {
      window.gameConsole?.log(
        `  ${world.worldId} — ${world.title || world.worldId} ` +
        `(${world.gridCount ?? '?'} grids, spawn ${world.spawnGrid || '?'})`
      )
    }
  }
}
