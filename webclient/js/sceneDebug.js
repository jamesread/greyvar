const STATUS_NAMES = [
  'PENDING', 'INIT', 'START', 'LOADING', 'CREATING',
  'RUNNING', 'PAUSED', 'SLEEPING', 'SHUTDOWN', 'DESTROYED'
]

export function sceneStatusName (status) {
  return STATUS_NAMES[status] ?? `unknown(${status})`
}

export function sceneSnapshot (label = '') {
  const mgr = window.phaser?.scene
  if (!mgr) {
    console.warn('[Greyvar] sceneSnapshot: phaser.scene missing')
    return null
  }

  const keys = ['splash', 'grid', 'hud', 'menu']
  const rows = keys.map((key) => {
    const scene = mgr.getScene(key)
    if (!scene) {
      return { key, exists: false }
    }

    const settings = scene.sys.settings
    return {
      key,
      exists: true,
      status: sceneStatusName(settings.status),
      statusCode: settings.status,
      active: settings.active,
      visible: settings.visible,
      isActive: mgr.isActive(key),
      isVisible: mgr.isVisible(key),
      isPaused: mgr.isPaused(key),
      isSleeping: mgr.isSleeping(key),
      index: mgr.getIndex(key),
      finished: scene.finished ?? null
    }
  })

  const info = {
    label: label || '(no label)',
    isProcessing: mgr.isProcessing,
    isBooted: mgr.isBooted,
    renderOrder: mgr.scenes?.map((s) => s.sys.settings.key) ?? [],
    gameState: {
      hackHasLoadedLogo: window.gameState?.hackHasLoadedLogo,
      hackHasLoadedInitialGrid: window.gameState?.hackHasLoadedInitialGrid,
      gridSceneSet: window.gameState?.gridScene != null,
      showGameScheduled: window.gameState?._showGamePending ?? false
    },
    serverOk: window.serverConnection?.isOk ?? false
  }

  console.group(`[Greyvar] scenes @ ${label || 'snapshot'}`)
  console.table(rows)
  console.log('context', info)
  console.groupEnd()

  return { rows, info }
}

export function logGreyvar (message, extra) {
  if (extra !== undefined) {
    console.log(`[Greyvar] ${message}`, extra)
  } else {
    console.log(`[Greyvar] ${message}`)
  }
}
