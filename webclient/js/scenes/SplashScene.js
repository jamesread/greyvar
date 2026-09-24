import Phaser from 'phaser'
import { logGreyvar, sceneSnapshot } from '../sceneDebug.js'

const COLORS = {
  bgTop: 0x1b1f3a,
  bgBottom: 0x0f1224,
  panel: 0x252b4a,
  panelBorder: 0x4ecca3,
  panelBorderDark: 0x2a8f6f,
  title: '#f5f7ff',
  subtitle: '#9aa4c7',
  label: '#c5cee0',
  pending: '#9aa4c7',
  ok: '#4ecca3',
  warn: '#e8b923',
  error: '#ff6b6b',
  accent: '#6c8cff'
}

export default class SplashScene extends Phaser.Scene {
  constructor () {
    super({ key: 'splash' })
    this.finished = false
    this.dotTimer = 0
  }

  preload () {
    this.load.setBaseURL('/res/')
    this.load.image('logo', 'img/logo32.png')
    window.backgroundMusic?.queueLoad(this)
  }

  create () {
    window.gameState.hackHasLoadedLogo = this.textures.exists('logo')
    logGreyvar('SplashScene.create()', { logoLoaded: window.gameState.hackHasLoadedLogo })
    sceneSnapshot('SplashScene.create')

    this.bg = this.add.graphics()
    this.panel = this.add.graphics()
    this.decor = this.add.graphics()

    this.logo = this.textures.exists('logo')
      ? this.add.image(0, 0, 'logo').setDisplaySize(96, 96)
      : null

    this.title = this.add.text(0, 0, 'Greyvar', {
      fontFamily: 'monospace',
      fontSize: '42px',
      color: COLORS.title,
      fontStyle: 'bold'
    }).setOrigin(0.5, 0)

    this.subtitle = this.add.text(0, 0, 'Collaborate to escape puzzle worlds', {
      fontFamily: 'monospace',
      fontSize: '14px',
      color: COLORS.subtitle
    }).setOrigin(0.5, 0)

    this.statusPanel = this.add.container(0, 0)

    this.lblServer = this.makeStatusRow('Server')
    this.lblResources = this.makeStatusRow('Resources')
    this.lblWorld = this.makeStatusRow('Game world')

    this.statusPanel.add([
      this.lblServer.row,
      this.lblResources.row,
      this.lblWorld.row
    ])

    this.hint = this.add.text(0, 0, 'Connecting...', {
      fontFamily: 'monospace',
      fontSize: '13px',
      color: COLORS.pending,
      align: 'center'
    }).setOrigin(0.5, 0)

    this.layout()
    this.scale.on('resize', this.layout, this)
  }

  makeStatusRow (label) {
    const row = this.add.container(0, 0)

    const dot = this.add.circle(0, 0, 5, 0x9aa4c7)
    const name = this.add.text(18, 0, label, {
      fontFamily: 'monospace',
      fontSize: '16px',
      color: COLORS.label
    }).setOrigin(0, 0.5)

    const value = this.add.text(0, 0, '...', {
      fontFamily: 'monospace',
      fontSize: '16px',
      color: COLORS.pending
    }).setOrigin(1, 0.5)

    row.add([dot, name, value])
    row.setSize(280, 24)

    return { row, dot, value }
  }

  setStatus (entry, text, color, dotColor) {
    entry.value.setText(text)
    entry.value.setColor(color)
    entry.dot.setFillStyle(dotColor)
  }

  layout () {
    const { width, height } = this.scale
    const cx = width / 2
    const pad = 36
    const panelW = Math.min(420, width - 48)
    const panelH = 380
    const panelX = cx - panelW / 2
    const panelY = Math.max(24, height / 2 - panelH / 2)
    const innerW = panelW - pad * 2

    this.redrawBackground(width, height)
    this.redrawPanel(panelX, panelY, panelW, panelH)

    let y = panelY + pad

    if (this.logo) {
      this.logo.setPosition(cx, y + 48)
      y += 108
    } else {
      y += 12
    }

    this.title.setPosition(cx, y)
    y += 52

    this.subtitle.setPosition(cx, y)
    y += 36

    this.decor.clear()
    this.decor.lineStyle(2, COLORS.accent, 0.35)
    this.decor.beginPath()
    this.decor.moveTo(panelX + pad, y)
    this.decor.lineTo(panelX + panelW - pad, y)
    this.decor.strokePath()
    y += 28

    const rowGap = 32
    const rows = [this.lblServer, this.lblResources, this.lblWorld]
    for (const entry of rows) {
      entry.row.setSize(innerW, 24)
      entry.value.setX(innerW)
    }

    this.lblServer.row.setPosition(panelX + pad, y)
    this.lblResources.row.setPosition(panelX + pad, y + rowGap)
    this.lblWorld.row.setPosition(panelX + pad, y + rowGap * 2)

    this.hint.setPosition(cx, panelY + panelH - pad)
  }

  redrawBackground (width, height) {
    this.bg.clear()
    this.bg.fillGradientStyle(COLORS.bgTop, COLORS.bgTop, COLORS.bgBottom, COLORS.bgBottom, 1)
    this.bg.fillRect(0, 0, width, height)

    this.bg.fillStyle(COLORS.accent, 0.04)
    for (let x = 0; x < width; x += 24) {
      for (let y = 0; y < height; y += 24) {
        if ((x + y) % 48 === 0) {
          this.bg.fillRect(x, y, 2, 2)
        }
      }
    }
  }

  redrawPanel (x, y, w, h) {
    this.panel.clear()

    this.panel.fillStyle(0x000000, 0.35)
    this.panel.fillRoundedRect(x + 6, y + 8, w, h, 16)

    this.panel.fillStyle(COLORS.panel, 0.96)
    this.panel.fillRoundedRect(x, y, w, h, 16)

    this.panel.lineStyle(3, COLORS.panelBorderDark, 1)
    this.panel.strokeRoundedRect(x + 2, y + 2, w - 4, h - 4, 14)

    this.panel.lineStyle(2, COLORS.panelBorder, 1)
    this.panel.strokeRoundedRect(x, y, w, h, 16)
  }

  update (_time, delta) {
    if (this.finished) {
      return
    }

    this.dotTimer += delta
    const dots = '.'.repeat(1 + Math.floor(this.dotTimer / 400) % 3)

    const serverOk = window.serverConnection.isOk
    const resourcesOk = window.gameState.hackHasLoadedLogo
    const resourcesLoading = this.load.isLoading()
    const worldOk = window.gameState.gridScene != null

    if (serverOk) {
      this.setStatus(this.lblServer, 'Connected', COLORS.ok, 0x4ecca3)
    } else {
      this.setStatus(this.lblServer, 'Waiting' + dots, COLORS.pending, 0x9aa4c7)
      this.hint.setText('Opening connection to server' + dots)
      return
    }

    if (resourcesOk) {
      this.setStatus(this.lblResources, 'Ready', COLORS.ok, 0x4ecca3)
    } else if (resourcesLoading) {
      this.setStatus(this.lblResources, 'Loading' + dots, COLORS.warn, 0xe8b923)
      this.hint.setText('Loading game assets' + dots)
      return
    } else {
      this.setStatus(this.lblResources, 'Failed', COLORS.error, 0xff6b6b)
      this.hint.setText('Could not load assets from /res')
      return
    }

    if (worldOk) {
      this.setStatus(this.lblWorld, 'Ready', COLORS.ok, 0x4ecca3)
      this.hint.setText('Starting game...')
      this.finished = true
      logGreyvar('SplashScene: world ready, waiting for finishShowGame()', {
        splashStatus: this.sys.settings.status,
        splashVisible: this.sys.settings.visible
      })
      sceneSnapshot('SplashScene world ready')
      return
    }

    this.setStatus(this.lblWorld, 'Loading' + dots, COLORS.warn, 0xe8b923)
    this.hint.setText('Waiting for game world' + dots)
  }

  shutdown () {
    logGreyvar('SplashScene.shutdown()')
    sceneSnapshot('SplashScene.shutdown')
    this.scale.off('resize', this.layout, this)
  }
}
