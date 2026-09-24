const STORAGE_KEY = 'greyvar.musicMuted'
const MUSIC_KEY = 'bgm_peaceful'
const MUSIC_PATH = 'snd/world/peaceful_loop.mp3'
const DEFAULT_VOLUME = 0.35

/**
 * Looping background music for the webclient.
 * Uses Phaser's HTML5 audio path (Web Audio is disabled in game config).
 */
export default class BackgroundMusic {
  constructor (game) {
    this.game = game
    this.sound = null
    this.unlocked = false
    this.ready = false
    this.muted = readMutedPreference()
    this._unlockBound = false
  }

  static init (game) {
    if (window.backgroundMusic == null) {
      window.backgroundMusic = new BackgroundMusic(game)
    }
    return window.backgroundMusic
  }

  /** Queue the track on a scene loader (call from preload). */
  queueLoad (scene) {
    if (scene.cache.audio.exists(MUSIC_KEY)) {
      this.ready = true
      return
    }

    scene.load.audio(MUSIC_KEY, MUSIC_PATH)
    scene.load.once(`filecomplete-audio-${MUSIC_KEY}`, () => {
      this.ready = true
      this.tryPlay()
    })
  }

  /** Bind one-shot unlock to the first user gesture (autoplay policy). */
  bindUnlock () {
    if (this._unlockBound) {
      return
    }
    this._unlockBound = true

    const unlock = () => {
      this.unlocked = true
      this.game.sound.unlock()
      this.tryPlay()
      window.removeEventListener('pointerdown', unlock, true)
      window.removeEventListener('keydown', unlock, true)
    }

    window.addEventListener('pointerdown', unlock, true)
    window.addEventListener('keydown', unlock, true)
  }

  tryPlay () {
    if (!this.ready || !this.unlocked || this.muted) {
      return
    }

    if (this.sound == null) {
      if (!this.game.cache.audio.exists(MUSIC_KEY)) {
        return
      }
      this.sound = this.game.sound.add(MUSIC_KEY, {
        loop: true,
        volume: DEFAULT_VOLUME
      })
    }

    if (!this.sound.isPlaying) {
      this.sound.play()
    }
  }

  isMuted () {
    return this.muted
  }

  setMuted (muted) {
    this.muted = Boolean(muted)
    writeMutedPreference(this.muted)

    if (this.sound == null) {
      if (!this.muted) {
        this.tryPlay()
      }
      return this.muted
    }

    if (this.muted) {
      this.sound.pause()
    } else {
      this.unlocked = true
      if (this.sound.isPaused) {
        this.sound.resume()
      } else {
        this.tryPlay()
      }
    }

    return this.muted
  }

  toggleMute () {
    return this.setMuted(!this.muted)
  }

  statusLine () {
    const playing = this.sound?.isPlaying ? 'playing' : (this.sound?.isPaused ? 'paused' : 'stopped')
    return `music: ${playing}, muted=${this.muted}, unlocked=${this.unlocked}, ready=${this.ready}`
  }
}

function readMutedPreference () {
  try {
    return window.localStorage.getItem(STORAGE_KEY) === '1'
  } catch {
    return false
  }
}

function writeMutedPreference (muted) {
  try {
    window.localStorage.setItem(STORAGE_KEY, muted ? '1' : '0')
  } catch {
    // ignore quota / private mode
  }
}
