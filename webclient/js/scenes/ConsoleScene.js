import Phaser from 'phaser'

export default class ConsoleScene extends Phaser.Scene {
  constructor () {
    super({ key: 'console' })
  }

  create () {
    window.gameConsole?.attachScene(this)
    this.events.once('shutdown', this.onShutdown, this)
  }

  onShutdown () {
    window.gameConsole?.detachScene()
  }
}
