// One clock for every layer of a rig: steps through a clip's {frame, ms}
// list so all layers always show the same pose.

export default class AnimationController {
  /** @param {Object<string, Array<{frame:number, ms:number, pose:string}>>} clips */
  constructor (clips) {
    this.clips = clips
    this.clipName = null
    this.steps = null
    this.index = 0
    this.elapsed = 0
    this.loop = true
    this.onComplete = null
  }

  get frame () {
    return this.steps?.[this.index]?.frame ?? 0
  }

  get pose () {
    return this.steps?.[this.index]?.pose ?? null
  }

  hasClip (name) {
    return Array.isArray(this.clips[name]) && this.clips[name].length > 0
  }

  /**
   * @param {string} name clip name
   * @param {{loop?: boolean, onComplete?: Function, restart?: boolean}} [opts]
   *   restart=false (default) keeps a looping clip running if already playing,
   *   so repeated move messages do not pin it to frame 0.
   * @returns {boolean} whether the clip exists
   */
  play (name, { loop = true, onComplete = null, restart = false } = {}) {
    if (!this.hasClip(name)) {
      return false
    }
    if (!restart && this.clipName === name && this.loop && loop) {
      return true
    }
    this.clipName = name
    this.steps = this.clips[name]
    this.index = 0
    this.elapsed = 0
    this.loop = loop
    this.onComplete = onComplete
    return true
  }

  /** Advance time. Returns true when the visible frame changed. */
  tick (deltaMs) {
    if (this.steps == null) {
      return false
    }
    const before = this.frame
    this.elapsed += deltaMs
    // Bounded loop: a huge delta (tab in background) just wraps.
    for (let guard = 0; guard < 1000; guard++) {
      const ms = Math.max(1, this.steps[this.index].ms)
      if (this.elapsed < ms) {
        break
      }
      this.elapsed -= ms
      if (this.index + 1 < this.steps.length) {
        this.index++
      } else if (this.loop) {
        this.index = 0
      } else {
        this.elapsed = 0
        const done = this.onComplete
        this.onComplete = null
        if (done != null) {
          done()
        }
        break
      }
    }
    return this.frame !== before
  }
}
