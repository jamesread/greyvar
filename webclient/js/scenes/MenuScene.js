import Phaser from 'phaser'

export default class MenuScene extends Phaser.Scene {
  constructor () {
    super({ key: 'menu' })
  }

  create () {
    window.menuSystem?.attachScene(this)
    this.events.once('shutdown', this.onShutdown, this)
  }

  onShutdown () {
    window.menuSystem?.detachScene()
  }
}
