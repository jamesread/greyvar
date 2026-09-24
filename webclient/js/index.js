import './movementDebug.js'
import GameState from './GameState.js'
import ServerConnection from './ServerConnection.js'
import GridScene from './scenes/GridScene.js'
import SplashScene from './scenes/SplashScene.js'
import HudScene from './scenes/HudScene.js'
import MenuScene from './scenes/MenuScene.js'
import MenuSystem, { registerDefaultMenus } from './menu/MenuSystem.js'
import GameConsole from './console/GameConsole.js'
import { registerDefaultCommands } from './console/commands.js'
import ConsoleScene from './scenes/ConsoleScene.js'
import BackgroundMusic from './BackgroundMusic.js'
import Phaser from 'phaser'
import { sceneSnapshot } from './sceneDebug.js'

export function init() {
  if (window.phaser) {
    console.warn('[Greyvar] init() called twice, ignoring')
    return
  }

  window.gameState = new GameState()

  window.phaser = new Phaser.Game({
    type: Phaser.WEBGL,
    width: 320,
    height: 256,
    audio: {
      disableWebAudio: true,
    },
    input: {
      gamepad: true,
    },
    background: 'orange',
    render: { // These settings apparently need to be set for Android: https://github.com/photonstorm/phaser/issues/5659
      batchSize: 1024,
      maxTextures: 7,
      pixelArt: true,
      antialias: false,
      roundPixels: true,
    },
    scale: {
      mode: Phaser.Scale.RESIZE,
      autoCenter: Phaser.Scale.CENTER_BOTH,
      zoom: Phaser.Scale.MAX_ZOOM, // real zoom is scaled at runtime
    },
    autoRound: true,
    physics: {
      default: 'arcade',
      arcade: {
        //debug: 'true'
      }
    }
  })

  window.phaser.scene.add('hud', HudScene, false)
  window.phaser.scene.add('menu', MenuScene, false)
  window.phaser.scene.add('console', ConsoleScene, false)
  window.phaser.scene.add('splash', SplashScene, true)

  const menuSystem = MenuSystem.init(window.phaser)
  registerDefaultMenus(menuSystem)

  const gameConsole = GameConsole.init(window.phaser)
  registerDefaultCommands(gameConsole)

  const backgroundMusic = BackgroundMusic.init(window.phaser)
  backgroundMusic.bindUnlock()

  const canvas = window.phaser.canvas
  canvas.setAttribute('tabindex', '0')
  canvas.addEventListener('pointerdown', () => canvas.focus())

  window.debugGreyvarScenes = sceneSnapshot
  console.log('[Greyvar] Debug: call window.debugGreyvarScenes("manual") in the console to inspect scene state')

  window.serverConnection = new ServerConnection()
}

window.resizeGame = () => {
  if (window.gameState?.gridScene == null) {
    return
  }

  window.gameState.gridScene.configureCamera()
  console.log('resized game')
}

window.addEventListener('resize', window.resizeGame, false)
window.addEventListener('orientationchange', window.resizeGame)
