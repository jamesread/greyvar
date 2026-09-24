import { dumpMovementDebug, dumpGameStatus } from '../movementDebug.js'
import { sceneSnapshot } from '../sceneDebug.js'
import { toggleCollisionDebug } from '../collisionDebug.js'

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
    .registerCommand('debug collision', () => toggleCollisionDebug(), {
      description: 'Toggle collision overlay (red=blocking rects/tiles, blue=player bounds)',
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
