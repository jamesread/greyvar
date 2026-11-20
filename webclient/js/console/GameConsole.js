const CONSOLE_THEME = {
  bg: 0x0a120a,
  bgAlpha: 0.88,
  text: '#7cfc7c',
  prompt: '#b0b0b0',
  fontFamily: 'Courier New, Courier, monospace',
  fontSize: '14px',
  lineHeight: 16,
  pad: 8,
  heightRatio: 0.5,
  maxLines: 300,
  promptText: ']'
}

export default class GameConsole {
  constructor (game) {
    this.game = game
    this.sceneKey = 'console'
    this.openState = false
    this.lines = []
    this.input = ''
    this.inputDraft = ''
    this.history = []
    this.historyIndex = -1
    this.scrollOffset = 0
    this.viewportLines = 10 // updated by layout()
    this.commands = new Map()
    this.boundScene = null
    this.ui = null
    this.globalKeyHandler = this.onGlobalKeydown.bind(this)
    document.addEventListener('keydown', this.globalKeyHandler, true)
  }

  static init (game) {
    if (window.gameConsole == null) {
      window.gameConsole = new GameConsole(game)
    }
    return window.gameConsole
  }

  registerCommand (name, handler, meta = {}) {
    this.commands.set(name.toLowerCase(), {
      name: name.toLowerCase(),
      handler,
      description: meta.description ?? '',
      usage: meta.usage ?? '',
      hidden: meta.hidden ?? false
    })
    return this
  }

  isOpen () {
    return this.openState
  }

  log (message) {
    const text = String(message)
    const added = text.split('\n')
    for (const line of added) {
      this.lines.push(line)
    }
    if (this.scrollOffset > 0) {
      this.scrollOffset += added.length // stay on what is being read
    }

    while (this.lines.length > CONSOLE_THEME.maxLines) {
      this.lines.shift()
    }

    if (this.openState) {
      this.render()
    }
  }

  clear () {
    this.lines = []
    this.scrollOffset = 0
    if (this.openState) {
      this.render()
    }
  }

  toggle () {
    if (this.isOpen()) {
      this.close()
    } else {
      this.open()
    }
  }

  open () {
    if (this.openState) {
      return
    }

    this.openState = true
    this.scrollOffset = 0
    this.historyIndex = -1
    this.inputDraft = this.input

    if (this.lines.length === 0) {
      this.log('Greyvar console — type "help" for commands')
    }

    this.ensureSceneRunning()
    this.render()
  }

  close () {
    if (!this.openState) {
      return
    }

    this.openState = false
    this.inputDraft = this.input

    if (this.game.scene.isActive(this.sceneKey)) {
      this.game.scene.stop(this.sceneKey)
    }

    this.clearUi()
    this.boundScene = null
  }

  ensureSceneRunning () {
    const mgr = this.game.scene
    if (!mgr.getScene(this.sceneKey)) {
      return
    }

    if (!mgr.isActive(this.sceneKey)) {
      mgr.run(this.sceneKey)
    }

    mgr.bringToTop(this.sceneKey)
  }

  attachScene (scene) {
    this.boundScene = scene
    scene.input.keyboard?.on('keydown', this.onSceneKeydown, this)
    scene.input.on('wheel', this.onWheel, this)
    this.render()
  }

  detachScene () {
    this.boundScene?.input.keyboard?.off('keydown', this.onSceneKeydown, this)
    this.boundScene?.input.off('wheel', this.onWheel, this)
    this.clearUi()
    this.boundScene = null
  }

  clearUi () {
    if (this.ui?.root != null) {
      this.ui.root.destroy(true)
    }
    this.ui = null
  }

  onGlobalKeydown (event) {
    if (event.code !== 'Backquote' || event.repeat) {
      return
    }

    if (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) {
      return
    }

    event.preventDefault()
    event.stopPropagation()
    this.toggle()
  }

  onWheel (_pointer, _over, _dx, dy) {
    if (this.openState && dy !== 0) {
      this.scrollBy(dy < 0 ? 3 : -3)
    }
  }

  onSceneKeydown (event) {
    if (!this.openState) {
      return
    }

    if (event.code === 'Backquote') {
      event.preventDefault()
      return
    }

    if (event.code === 'Escape') {
      event.preventDefault()
      this.close()
      return
    }

    if (event.code === 'Enter') {
      event.preventDefault()
      this.submitInput()
      return
    }

    if (event.code === 'Backspace') {
      // Ctrl+Backspace clears the whole line (and keeps the browser from
      // acting on it); plain Backspace deletes one character.
      event.preventDefault()
      this.input = event.ctrlKey || event.metaKey ? '' : this.input.slice(0, -1)
      this.historyIndex = -1
      this.render()
      return
    }

    // Scrollback: PageUp/PageDown by a page, Shift+Up/Down by a line,
    // Ctrl+Home/End to the top/bottom. Plain Up/Down walk the history.
    const page = Math.max(1, this.viewportLines - 1)
    const scroll = {
      PageUp: page,
      PageDown: -page,
      ArrowUp: event.shiftKey ? 1 : null,
      ArrowDown: event.shiftKey ? -1 : null,
      Home: event.ctrlKey ? Infinity : null,
      End: event.ctrlKey ? -Infinity : null
    }[event.code]
    if (scroll != null) {
      event.preventDefault()
      this.scrollBy(scroll)
      return
    }

    if (event.code === 'ArrowUp') {
      event.preventDefault()
      this.historyUp()
      return
    }

    if (event.code === 'ArrowDown') {
      event.preventDefault()
      this.historyDown()
      return
    }

    if (event.key.length === 1 && !event.ctrlKey && !event.metaKey && !event.altKey) {
      event.preventDefault()
      this.input += event.key
      this.render()
    }
  }

  /** Lines the scrollback can move up (0 when it all fits). */
  maxScroll () {
    return Math.max(0, this.lines.length - this.viewportLines)
  }

  /** Scroll the output: positive is back (older lines), negative forward. */
  scrollBy (n) {
    const next = Math.min(this.maxScroll(), Math.max(0, this.scrollOffset + n))
    if (next !== this.scrollOffset) {
      this.scrollOffset = next
      this.render()
    }
  }

  historyUp () {
    if (this.history.length === 0) {
      return
    }

    if (this.historyIndex === -1) {
      this.inputDraft = this.input
    }

    const nextIndex = this.historyIndex === -1
      ? this.history.length - 1
      : Math.max(0, this.historyIndex - 1)

    this.historyIndex = nextIndex
    this.input = this.history[nextIndex]
    this.render()
  }

  historyDown () {
    if (this.historyIndex === -1) {
      return
    }

    const nextIndex = this.historyIndex + 1
    if (nextIndex >= this.history.length) {
      this.historyIndex = -1
      this.input = this.inputDraft
    } else {
      this.historyIndex = nextIndex
      this.input = this.history[nextIndex]
    }

    this.render()
  }

  submitInput () {
    const line = this.input.trim()
    if (line !== '') {
      this.log(`${CONSOLE_THEME.promptText}${line}`)
      this.history.push(line)
      this.historyIndex = -1
      this.inputDraft = ''
      this.execute(line)
    }

    this.input = ''
    this.scrollOffset = 0
    this.render()
  }

  execute (line) {
    const parts = line.trim().split(/\s+/).filter(Boolean)

    if (parts.length === 0) {
      return
    }

    let cmd = null
    let args = []

    for (let len = parts.length; len >= 1; len--) {
      const name = parts.slice(0, len).join(' ').toLowerCase()
      if (this.commands.has(name)) {
        cmd = this.commands.get(name)
        args = parts.slice(len)
        break
      }
    }

    if (cmd == null) {
      this.log(`unknown command: ${parts[0]}`)
      return
    }

    try {
      const result = cmd.handler(args, {
        console: this,
        game: this.game,
        args
      })

      if (result == null) {
        return
      }

      if (result instanceof Promise) {
        result.then((resolved) => {
          if (resolved == null) {
            return
          }
          if (Array.isArray(resolved)) {
            for (const row of resolved) {
              this.log(row)
            }
          } else {
            this.log(String(resolved))
          }
        }).catch((err) => {
          this.log(`error: ${err.message ?? err}`)
        })
        return
      }

      if (Array.isArray(result)) {
        for (const row of result) {
          this.log(row)
        }
      } else {
        this.log(String(result))
      }
    } catch (err) {
      this.log(`error: ${err.message ?? err}`)
    }
  }

  visibleLines (viewportLines) {
    const total = this.lines.length
    const maxScroll = Math.max(0, total - viewportLines)
    const offset = Math.min(this.scrollOffset, maxScroll)
    this.scrollOffset = offset
    const end = total - offset
    const start = Math.max(0, end - viewportLines)
    return this.lines.slice(start, end)
  }

  render () {
    const scene = this.boundScene
    if (scene == null || !this.openState) {
      return
    }

    this.clearUi()

    const root = scene.add.container(0, 0)
    root.setScrollFactor(0)
    root.setDepth(2000)

    const bg = scene.add.rectangle(0, 0, 10, 10, CONSOLE_THEME.bg, CONSOLE_THEME.bgAlpha)
    bg.setOrigin(0)

    const outputText = scene.add.text(0, 0, '', {
      fontFamily: CONSOLE_THEME.fontFamily,
      fontSize: CONSOLE_THEME.fontSize,
      color: CONSOLE_THEME.text,
      lineSpacing: 2
    }).setOrigin(0)

    const inputText = scene.add.text(0, 0, '', {
      fontFamily: CONSOLE_THEME.fontFamily,
      fontSize: CONSOLE_THEME.fontSize,
      color: CONSOLE_THEME.text
    }).setOrigin(0)

    root.add([bg, outputText, inputText])

    const layout = () => {
      const { width, height } = scene.scale
      const pad = CONSOLE_THEME.pad
      const consoleH = Math.floor(height * CONSOLE_THEME.heightRatio)
      const lineH = CONSOLE_THEME.lineHeight
      const viewportLines = Math.max(1, Math.floor((consoleH - pad * 2 - lineH) / lineH))

      bg.setPosition(0, 0)
      bg.setSize(width, consoleH)

      this.viewportLines = viewportLines
      const visible = this.visibleLines(viewportLines)
      if (this.scrollOffset > 0) {
        // Show that newer lines are below, and how far back we are.
        visible[visible.length - 1] = `-- scrolled back ${this.scrollOffset} line(s); PageDown or Ctrl+End for latest --`
      }
      outputText.setText(visible.join('\n'))
      outputText.setPosition(pad, pad)
      outputText.setWordWrapWidth(width - pad * 2)

      inputText.setText(`${CONSOLE_THEME.promptText}${this.input}_`)
      inputText.setPosition(pad, consoleH - pad - lineH)
    }

    layout()
    scene.scale.off('resize', this.ui?.layoutHandler)
    this.ui = {
      root,
      layout,
      layoutHandler: layout
    }
    scene.scale.on('resize', layout)
  }
}

export { CONSOLE_THEME }
