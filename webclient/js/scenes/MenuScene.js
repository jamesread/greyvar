import Phaser from 'phaser'
import { preloadMenuBorder } from '../menu/menuBorder.js'

export default class MenuScene extends Phaser.Scene {
  constructor () {
    super({ key: 'menu' })
  }

  preload () {
    this.load.setBaseURL('/res/')
    preloadMenuBorder(this.load)
  }

  create () {
    window.menuSystem?.attachScene(this)
    this.events.once('shutdown', this.onShutdown, this)
  }

  onShutdown () {
    window.menuSystem?.detachScene()
  }
}
