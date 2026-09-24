export function movementPollContext () {
  const keyboard = window.phaser?.scene?.getScene('grid')?.input?.keyboard
    ?? window.phaser?.scene?.getScene('hud')?.input?.keyboard

  return {
    keyboardReady: keyboard != null,
    keyboardEnabled: keyboard?.enabled ?? false,
    serverOk: window.serverConnection?.isOk ?? false,
    wsState: window.serverConnection?.sock?.readyState,
    gridActive: window.phaser?.scene?.isActive('grid') ?? false,
    hudActive: window.phaser?.scene?.isActive('hud') ?? false,
    gridSceneSet: window.gameState?.gridScene != null,
    localPlayerEntityId: window.gameState?.localPlayerEntityId ?? null,
    canvasFocused: document.activeElement === window.phaser?.canvas,
  }
}

export function dumpMovementDebug () {
  const ctx = movementPollContext()
  const grid = window.phaser?.scene?.getScene('grid')
  const localId = window.gameState?.localPlayerEntityId
  const localEnt = localId != null ? grid?.entities?.[localId] : null

  const snapshot = {
    ...ctx,
    localPlayer: localEnt == null ? null : {
      entityId: localEnt.entityId,
      x: localEnt.x,
      y: localEnt.y,
      hasImg: localEnt.img != null,
      spriteX: localEnt.img?.x,
      spriteY: localEnt.img?.y,
    },
    playerEntityCount: grid
      ? Object.values(grid.entities).filter((e) => e.definition === 'player').length
      : 0,
  }

  console.log('[Greyvar] dumpMovementDebug', snapshot)
  return snapshot
}

export function buildGameStatus () {
  const grid = window.gameState?.grid
  const gridScene = window.gameState?.gridScene
  const localId = window.gameState?.localPlayerEntityId
  const localEnt = localId != null ? gridScene?.entities?.[localId] : null

  const x = localEnt?.x ?? localEnt?.img?.x ?? null
  const y = localEnt?.y ?? localEnt?.img?.y ?? null

  return {
    connected: window.serverConnection?.isOk ?? false,
    wsState: window.serverConnection?.sock?.readyState ?? null,
    worldId: grid?.worldId ?? null,
    gridId: grid?.gridId ?? null,
    gridLoaded: window.gameState?.hackHasLoadedInitialGrid ?? false,
    localPlayerEntityId: localId ?? null,
    x,
    y,
    tileRow: y != null ? Math.trunc(y / 16) : null,
    tileCol: x != null ? Math.trunc(x / 16) : null
  }
}

export function dumpGameStatus () {
  const status = buildGameStatus()
  console.log('[Greyvar] status', status)
  return status
}

window.dumpMovementDebug = dumpMovementDebug
window.dumpGameStatus = dumpGameStatus
