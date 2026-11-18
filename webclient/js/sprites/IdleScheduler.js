// Idle behaviour from rig.json `idle`: loop the base clip, and every
// intervalMs[min..max] play one weighted-random variant once, then go back.

export function pickWeighted (variants, rng = Math.random) {
  const total = variants.reduce((sum, v) => sum + Math.max(0, v.weight), 0)
  if (total <= 0) {
    return null
  }
  let roll = rng() * total
  for (const v of variants) {
    roll -= Math.max(0, v.weight)
    if (roll < 0) {
      return v
    }
  }
  return variants[variants.length - 1]
}

export default class IdleScheduler {
  /**
   * @param {{base: string, variants?: Array<{clip:string, weight:number}>, intervalMs?: number[]}} spec
   * @param {import('./AnimationController.js').default} controller
   * @param {() => number} [rng]
   */
  constructor (spec, controller, rng = Math.random) {
    this.spec = spec ?? {}
    this.controller = controller
    this.rng = rng
    this.active = false
    this.inVariant = false
    this.untilNext = 0
  }

  nextInterval () {
    const [min, max] = this.spec.intervalMs ?? [0, 0]
    return min + this.rng() * Math.max(0, max - min)
  }

  start () {
    if (this.active) {
      return
    }
    this.active = true
    this.playBase()
  }

  stop () {
    this.active = false
    this.inVariant = false
  }

  playBase () {
    this.inVariant = false
    this.untilNext = this.nextInterval()
    this.controller.play(this.spec.base, { loop: true, restart: true })
  }

  tick (deltaMs) {
    if (!this.active || this.inVariant) {
      return
    }
    const variants = (this.spec.variants ?? []).filter((v) => this.controller.hasClip(v.clip))
    if (variants.length === 0) {
      return
    }
    this.untilNext -= deltaMs
    if (this.untilNext > 0) {
      return
    }
    const pick = pickWeighted(variants, this.rng)
    if (pick == null) {
      this.untilNext = this.nextInterval()
      return
    }
    this.inVariant = true
    this.controller.play(pick.clip, {
      loop: false,
      restart: true,
      onComplete: () => {
        if (this.active) {
          this.playBase()
        }
      }
    })
  }
}
