import { dumpMovementDebug, dumpGameStatus } from '../movementDebug.js'
import { sceneSnapshot } from '../sceneDebug.js'
import { toggleCollisionDebug } from '../collisionDebug.js'
import { toggleDebugPos } from '../debugPos.js'
import { getRig, sizeOf, RIG_URLS } from '../sprites/PlayerRig.js'
import { parseAppearanceArgs, describeAppearanceOptions } from '../sprites/appearance.js'
import { parseRiftArgs } from './riftArgs.js'
import { parseZoomArg, DEFAULT_ZOOM, MIN_ZOOM, MAX_ZOOM } from '../camera/zoom.js'

export function registerDefaultCommands (gameConsole) {
  gameConsole
    .registerCommand('help', () => {
      const rows = ['Available commands:']
      const names = [...gameConsole.commands.keys()].sort()
      for (const name of names) {
        const cmd = gameConsole.commands.get(name)
        if (cmd.hidden) {
          continue
        }
        const usage = cmd.usage ? ` ${cmd.usage}` : ''
        rows.push(`  ${cmd.name}${usage} — ${cmd.description || 'no description'}`)
      }
      return rows
    }, {
      description: 'List console commands'
    })
    .registerCommand('clear', (_args, ctx) => {
      ctx.console.clear()
    }, {
      description: 'Clear console output'
    })
    .registerCommand('echo', (args) => args.join(' '), {
      description: 'Print text',
      usage: '<text...>'
    })
    .registerCommand('version', () => {
      const response = window.gameState?.messages?.find?.((m) => m.serverVersion != null)
      return response?.serverVersion != null
        ? `server version: ${response.serverVersion}`
        : 'server version: unknown (not connected yet)'
    }, {
      description: 'Show server version'
    })
    .registerCommand('status', () => {
      const s = dumpGameStatus()
      return [
        `connected: ${s.connected}`,
        `websocket readyState: ${s.wsState ?? 'n/a'}`,
        `world: ${s.worldId ?? 'n/a'}`,
        `grid: ${s.gridId ?? 'n/a'}`,
        `local player entity: ${s.localPlayerEntityId ?? 'n/a'}`,
        `position: (${s.x ?? '?'}, ${s.y ?? '?'})`,
        `tile: row ${s.tileRow ?? '?'}, col ${s.tileCol ?? '?'}`,
        `grid loaded: ${s.gridLoaded}`
      ]
    }, {
      description: 'Show connection and game status'
    })
    .registerCommand('movement', () => {
      const snapshot = dumpMovementDebug()
      return [
        `keyboard: ${snapshot.keyboardReady ? 'ready' : 'missing'} (${snapshot.keyboardEnabled ? 'enabled' : 'disabled'})`,
        `canvas focused: ${snapshot.canvasFocused}`,
        `local player: ${snapshot.localPlayer?.entityId ?? 'n/a'} @ (${snapshot.localPlayer?.x ?? '?'}, ${snapshot.localPlayer?.y ?? '?'})`,
        `player entities on grid: ${snapshot.playerEntityCount}`
      ]
    }, {
      description: 'Dump movement debug info',
      usage: ''
    })
    .registerCommand('scenes', () => {
      const snap = sceneSnapshot('console')
      if (snap == null) {
        return 'scene manager unavailable'
      }
      return snap.rows.map((row) => {
        if (!row.exists) {
          return `${row.key}: missing`
        }
        return `${row.key}: ${row.status} active=${row.isActive} visible=${row.isVisible}`
      })
    }, {
      description: 'List Phaser scene states'
    })
    .registerCommand('fullscreen', () => {
      if (!window.phaser.scale.fullscreen.active) {
        window.phaser.scale.startFullscreen()
        return 'fullscreen enabled'
      }
      window.phaser.scale.stopFullscreen()
      return 'fullscreen disabled'
    }, {
      description: 'Toggle fullscreen'
    })
    .registerCommand('reload', () => {
      window.location.reload()
      return 'reloading…'
    }, {
      description: 'Reload the page'
    })
    .registerCommand('world-list', () => {
      if (!window.serverConnection?.isOk) {
        return 'not connected to server'
      }

      window.serverConnection.sendWorldListRequest()
      return 'requesting world list from server...'
    }, {
      description: 'List available worlds',
      usage: ''
    })
    .registerCommand('world-load', (args) => {
      const worldId = args[0]
      if (worldId == null || worldId === '') {
        return 'usage: world-load <worldId>'
      }

      if (!window.serverConnection?.isOk) {
        return 'not connected to server'
      }

      window.serverConnection.sendWorldLoadRequest(worldId)
      return `requesting world load: ${worldId}...`
    }, {
      description: 'Load a world and teleport to its spawn grid',
      usage: '<worldId>'
    })
    .registerCommand('appearance', (args) => {
      const scene = window.gameState?.gridScene
      const gs = window.gameState
      const look = scene?.entities?.[gs?.localPlayerEntityId]?.appearance
      // size=adult|child switches rig; the other keys are checked against it.
      const sizeArg = args.find((a) => String(a).startsWith('size='))
      const size = sizeArg != null ? sizeArg.slice(5) : sizeOf(look)
      if (!Object.hasOwn(RIG_URLS, size)) {
        return `unknown size "${size}" (have: ${Object.keys(RIG_URLS).join(', ')})`
      }
      const rig = scene != null ? getRig(scene, size) : null
      if (rig == null) {
        return 'player rig not loaded yet'
      }
      if (args.length === 0) {
        return ['Usage: appearance key=value ...', `  size: ${Object.keys(RIG_URLS).join(', ')}`, ...describeAppearanceOptions(rig)]
      }
      const { request, errors } = parseAppearanceArgs(rig, args.filter((a) => a !== sizeArg))
      if (sizeArg != null) {
        request.rig = size
      }
      if (errors.length > 0) {
        return errors
      }
      if (!window.serverConnection?.isOk) {
        return 'not connected to server'
      }
      window.serverConnection.sendAppearanceRequest(request)
      return 'requesting appearance change...'
    }, {
      description: 'Change your look, e.g. appearance hair=blonde shirt=green held=key size=child (no args lists options)',
      usage: '[key=value...]'
    })
    .registerCommand('rift', (args) => {
      const { request, errors } = parseRiftArgs(args)
      if (errors.length > 0) {
        return [...errors, 'usage: rift [level] [seed=N] [mechanics=lever,keyChest,pushBlock,timedSwitch] | rift leave']
      }
      if (!window.serverConnection?.isOk) {
        return 'not connected to server'
      }
      window.serverConnection.sendRiftRequest(request)
      if (request.leave) {
        return 'fleeing the rift...'
      }
      return request.level > 0 ? `opening rift level ${request.level}...` : 'opening your next rift...'
    }, {
      description: 'Enter a rift, a generated puzzle world (default: one level above your best), or `rift leave`',
      usage: '[level] [seed=N] [mechanics=a,b] | leave'
    })
    .registerCommand('world-reload', (args) => {
      if (!window.serverConnection?.isOk) {
        return 'not connected to server'
      }

      const worldId = args[0]
      window.serverConnection.sendWorldReloadRequest(worldId)
      if (worldId != null && worldId !== '') {
        return `requesting world reload: ${worldId}...`
      }
      return 'requesting reload of current world...'
    }, {
      description: 'Reload a world from disk and return players to its spawn',
      usage: '[worldId]'
    })
    .registerCommand('debug collision', () => toggleCollisionDebug(), {
      description: 'Toggle collision overlay (red=blocking rects/tiles, blue=player bounds)',
      usage: ''
    })
    .registerCommand('zoom', (args) => {
      const scene = window.gameState?.gridScene
      if (scene?.cameras?.main == null) {
        return 'game not loaded yet'
      }
      const current = scene.zoomLevel ?? scene.cameras.main.zoom
      if (args.length === 0) {
        return `zoom ${current} (default ${DEFAULT_ZOOM}, range ${MIN_ZOOM}-${MAX_ZOOM}; higher is closer)`
      }
      const { zoom, error } = parseZoomArg(args[0], current)
      if (error != null) {
        return error
      }
      return `zoom ${scene.setZoomLevel(zoom)}`
    }, {
      description: 'Show or set the camera zoom: zoom 4, zoom in, zoom out, zoom +2, zoom reset (lower shows more of the map)',
      usage: '[N|in|out|+N|-N|reset]'
    })
    .registerCommand('debug pos', () => toggleDebugPos(), {
      description: "Toggle the player's position (pixels, tile, map) in the bottom-left corner",
      usage: ''
    })
    .registerCommand('music', (args) => {
      const bgm = window.backgroundMusic
      if (bgm == null) {
        return 'background music not initialized'
      }

      const cmd = (args[0] ?? 'status').toLowerCase()
      if (cmd === 'mute' || cmd === 'off') {
        bgm.setMuted(true)
        return bgm.statusLine()
      }
      if (cmd === 'unmute' || cmd === 'on') {
        bgm.setMuted(false)
        return bgm.statusLine()
      }
      if (cmd === 'toggle') {
        bgm.toggleMute()
        return bgm.statusLine()
      }
      return bgm.statusLine()
    }, {
      description: 'Background music status / mute control',
      usage: '[status|mute|unmute|toggle]'
    })
}
