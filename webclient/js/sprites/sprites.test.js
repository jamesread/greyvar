import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { join } from 'node:path'
import AnimationController from './AnimationController.js'
import IdleScheduler, { pickWeighted } from './IdleScheduler.js'
import { resolveAppearance, recolourMap, recolourPixels, recolouredTextureKey } from './appearance.js'

const resDir = fileURLToPath(new URL('../../../res/', import.meta.url))
const playerRig = JSON.parse(readFileSync(join(resDir, 'img/textures/entities/player/player.rig.json'), 'utf8'))

const clips = {
  idle: [{ pose: 'a', frame: 0, ms: 100 }, { pose: 'b', frame: 1, ms: 50 }],
  shrug: [{ pose: 'c', frame: 2, ms: 30 }, { pose: 'd', frame: 3, ms: 30 }],
  walk: [{ pose: 'w0', frame: 4, ms: 10 }, { pose: 'w1', frame: 5, ms: 10 }]
}

// Deterministic RNG cycling through fixed values.
function seq (...values) {
  let i = 0
  return () => values[i++ % values.length]
}

describe('AnimationController', () => {
  it('steps through frames by ms and loops', () => {
    const c = new AnimationController(clips)
    c.play('idle')
    expect(c.frame).toBe(0)
    expect(c.tick(99)).toBe(false)
    expect(c.tick(1)).toBe(true)
    expect(c.frame).toBe(1)
    c.tick(50)
    expect(c.frame).toBe(0)
  })

  it('does not restart a looping clip that is already playing', () => {
    const c = new AnimationController(clips)
    c.play('walk')
    c.tick(10)
    expect(c.frame).toBe(5)
    c.play('walk')
    expect(c.frame).toBe(5)
  })

  it('fires onComplete once for one-shot clips', () => {
    const c = new AnimationController(clips)
    let done = 0
    c.play('shrug', { loop: false, onComplete: () => done++ })
    c.tick(1000)
    c.tick(1000)
    expect(done).toBe(1)
  })

  it('rejects unknown clips', () => {
    expect(new AnimationController(clips).play('nope')).toBe(false)
  })
})

describe('IdleScheduler', () => {
  it('picks weighted variants', () => {
    const v = [{ clip: 'a', weight: 1 }, { clip: 'b', weight: 3 }]
    expect(pickWeighted(v, () => 0.0).clip).toBe('a')
    expect(pickWeighted(v, () => 0.3).clip).toBe('b')
    expect(pickWeighted(v, () => 0.99).clip).toBe('b')
    expect(pickWeighted([{ clip: 'a', weight: 0 }], () => 0.5)).toBe(null)
  })

  it('loops base, plays a variant after the interval, then returns', () => {
    const c = new AnimationController(clips)
    const idle = new IdleScheduler({ base: 'idle', variants: [{ clip: 'shrug', weight: 1 }], intervalMs: [500, 500] }, c, seq(0))
    idle.start()
    expect(c.clipName).toBe('idle')
    idle.tick(499)
    expect(c.clipName).toBe('idle')
    idle.tick(1)
    expect(c.clipName).toBe('shrug')
    c.tick(60) // shrug finishes
    expect(c.clipName).toBe('idle')
    expect(idle.inVariant).toBe(false)
  })

  it('stays on base when there are no variants', () => {
    const c = new AnimationController(clips)
    const idle = new IdleScheduler({ base: 'idle' }, c)
    idle.start()
    idle.tick(1e6)
    expect(c.clipName).toBe('idle')
  })
})

describe('appearance', () => {
  const rig = {
    layers: [{ name: 'hair', default: 'short', variants: ['short', 'long'], slots: ['hair'] }],
    slotOrder: ['skin', 'hair'],
    slots: { skin: ['#100000', '#200000'], hair: ['#000010', '#000020'] },
    palettes: {
      skin: { light: ['#100000', '#200000'], dark: ['#110000', '#220000'] },
      hair: { brown: ['#000010', '#000020'], blonde: ['#ffff00', '#ffff80'] }
    },
    defaults: { palettes: { skin: 'light', hair: 'brown' } }
  }

  it('fills defaults and ignores unknown values', () => {
    const a = resolveAppearance(rig, { parts: { hair: 'mohawk' }, palettes: { hair: 'blonde', skin: 'green' } })
    expect(a.parts.hair).toBe('short')
    expect(a.palettes).toEqual({ skin: 'light', hair: 'blonde' })
  })

  it('maps only non-identity key colours for the requested slots', () => {
    const m = recolourMap(rig, { skin: 'light', hair: 'blonde' }, ['skin', 'hair'])
    expect(m.size).toBe(2)
    expect(m.get(0x000010)).toBe(0xffff00)
    expect(recolourMap(rig, { hair: 'blonde' }, ['skin']).size).toBe(0)
  })

  it('recolours RGBA data and skips transparent pixels', () => {
    const data = new Uint8ClampedArray([0, 0, 0x10, 255, 0, 0, 0x10, 0, 1, 2, 3, 255])
    const n = recolourPixels(data, new Map([[0x000010, 0xffff00]]))
    expect(n).toBe(1)
    expect(Array.from(data.slice(0, 4))).toEqual([255, 255, 0, 255])
    expect(Array.from(data.slice(4, 8))).toEqual([0, 0, 0x10, 0])
  })

  it('keys textures only by slots the layer uses', () => {
    expect(recolouredTextureKey('hair.png', ['hair'], { skin: 'dark', hair: 'red' })).toBe('hair.png|hair=red')
    expect(recolouredTextureKey('x.png', [], { hair: 'red' })).toBe('x.png')
  })
})

// Guards the generated manifest the game actually loads.
describe('player.rig.json', () => {
  it('references sheets that exist', () => {
    for (const layer of playerRig.layers) {
      for (const path of Object.values(layer.sheets)) {
        expect(existsSync(join(resDir, path)), path).toBe(true)
      }
    }
  })

  it('has every clip frame in range and every state mapped', () => {
    for (const [name, steps] of Object.entries(playerRig.clips)) {
      for (const s of steps) {
        expect(s.frame, name).toBeLessThan(playerRig.frameCount)
        expect(s.ms, name).toBeGreaterThan(0)
      }
    }
    for (const clip of Object.values(playerRig.states)) {
      expect(playerRig.clips[clip]).toBeDefined()
    }
    expect(playerRig.clips[playerRig.idle.base]).toBeDefined()
  })

  it('resolves the default appearance to itself', () => {
    const a = resolveAppearance(playerRig, {})
    expect(a.palettes).toEqual(playerRig.defaults.palettes)
  })
})

describe('parseAppearanceArgs', () => {
  const rig = {
    layers: [{ name: 'hair', default: 'short', variants: ['short', 'long'] }, { name: 'body', default: 'default', variants: ['default'] }],
    slotOrder: ['hair', 'shirt'],
    palettes: { hair: { brown: [], blonde: [] }, shirt: { red: [] } }
  }

  it('splits palettes, parts and held item', async () => {
    const { parseAppearanceArgs } = await import('./appearance.js')
    const { request, errors } = parseAppearanceArgs(rig, ['hair=blonde', 'hair=long', 'shirt=red', 'held=key'])
    expect(errors).toEqual([])
    expect(request).toEqual({ parts: { hair: 'long' }, palettes: { hair: 'blonde', shirt: 'red' }, heldItem: 'key' })
  })

  it('reports unknown values with the options', async () => {
    const { parseAppearanceArgs } = await import('./appearance.js')
    const { errors } = parseAppearanceArgs(rig, ['hair=plaid', 'cape=red', 'oops'])
    expect(errors[0]).toContain('brown, blonde, short, long')
    expect(errors[1]).toBe('unknown key "cape"')
    expect(errors[2]).toContain('want key=value')
  })
})

describe('facing', () => {
  it('derives facing from walk/bump states', async () => {
    const { facingFromState } = await import('./facing.js')
    expect(facingFromState('walkUpLeft')).toBe('UpLeft')
    expect(facingFromState('bumpRight')).toBe('Right')
    expect(facingFromState('walkDown')).toBe('Down')
    expect(facingFromState('idle')).toBe(null)
    expect(facingFromState('walk')).toBe(null)
  })

  it('picks idle<Facing>, falls back to the horizontal part, else front idle', async () => {
    const { idleClipForFacing } = await import('./facing.js')
    const clips = { idleLeft: [{}], idleRight: [{}], idleUp: [{}] }
    expect(idleClipForFacing(clips, 'Left')).toBe('idleLeft')
    expect(idleClipForFacing(clips, 'Up')).toBe('idleUp')
    expect(idleClipForFacing(clips, 'UpRight')).toBe('idleRight')
    expect(idleClipForFacing(clips, 'DownLeft')).toBe('idleLeft')
    expect(idleClipForFacing(clips, 'Down')).toBe(null)
    expect(idleClipForFacing(clips, null)).toBe(null)
    expect(idleClipForFacing({}, 'Left')).toBe(null)
  })

  it('the player rig has a facing idle for every walk direction except down', () => {
    return import('./facing.js').then(({ idleClipForFacing }) => {
      for (const f of ['Left', 'Right', 'Up', 'UpLeft', 'UpRight', 'DownLeft', 'DownRight']) {
        expect(idleClipForFacing(playerRig.clips, f), f).not.toBe(null)
      }
    })
  })
})

describe('use animation', () => {
  it('picks use<Facing>, the side clip for diagonals, else the front', async () => {
    const { useClipForFacing, facingFromState } = await import('./facing.js')
    const clips = { useLeft: [{}], useRight: [{}], useUp: [{}], useDown: [{}] }
    expect(useClipForFacing(clips, 'Up')).toBe('useUp')
    expect(useClipForFacing(clips, 'UpLeft')).toBe('useLeft')
    expect(useClipForFacing(clips, 'DownRight')).toBe('useRight')
    expect(useClipForFacing(clips, null)).toBe('useDown')
    expect(useClipForFacing({}, 'Left')).toBe(null)
    expect(facingFromState('useUpLeft')).toBe('UpLeft')
  })

  it('the player rig has a one-shot use clip for every facing', async () => {
    const { useClipForFacing } = await import('./facing.js')
    for (const f of ['Up', 'Down', 'Left', 'Right', 'UpLeft', 'UpRight', 'DownLeft', 'DownRight']) {
      const clip = useClipForFacing(playerRig.clips, f)
      expect(clip, f).not.toBe(null)
      const total = playerRig.clips[clip].reduce((ms, step) => ms + step.ms, 0)
      expect(total, `${clip} should be a quick gesture`).toBeLessThan(600)
    }
  })

  it('plays once and then returns', () => {
    const c = new AnimationController({ useLeft: [{ frame: 1, ms: 60 }, { frame: 2, ms: 200 }], idle: [{ frame: 0, ms: 1000 }] })
    let done = false
    c.play('useLeft', { loop: false, restart: true, onComplete: () => { done = true } })
    c.tick(100)
    expect(c.frame).toBe(2)
    c.tick(300)
    expect(done).toBe(true)
  })
})

describe('applyWear (live preview of a dialog option)', () => {
  it('applies styles and colours; hair decides by value', async () => {
    const { applyWear, resolveAppearance } = await import('./appearance.js')
    const look = resolveAppearance(playerRig, { parts: { hair: 'short' }, palettes: { hair: 'brown', shirt: 'magenta' } })
    expect(applyWear(playerRig, look, { shirt: 'blue' }).palettes.shirt).toBe('blue')
    const longHair = applyWear(playerRig, look, { hair: 'long' })
    expect([longHair.parts.hair, longHair.palettes.hair]).toEqual(['long', 'brown'])
    const blonde = applyWear(playerRig, look, { hair: 'blonde' })
    expect([blonde.parts.hair, blonde.palettes.hair]).toEqual(['short', 'blonde'])
    expect(applyWear(playerRig, look, { shirt: 'plaid' }).palettes.shirt).toBe('magenta')
    expect(applyWear(playerRig, look, null)).toEqual(look)
    expect(look.palettes.shirt).toBe('magenta') // not modified
  })
})

describe('appearanceStore', () => {
  const mem = () => { const m = new Map(); return { getItem: (k) => m.get(k) ?? null, setItem: (k, v) => m.set(k, v) } }

  it('round-trips a saved look and builds a restore request', async () => {
    const { saveAppearance, loadSavedAppearance, restoreRequest } = await import('./appearanceStore.js')
    const s = mem()
    expect(loadSavedAppearance(s)).toBe(null)
    saveAppearance({ rig: 'adult', parts: { hair: 'long' }, palettes: { hair: 'red' }, heldItem: '' }, s)
    const saved = loadSavedAppearance(s)
    expect(saved).toEqual({ rig: 'adult', parts: { hair: 'long' }, palettes: { hair: 'red' }, heldItem: '' })
    expect(restoreRequest(saved).heldItem).toBe('-')
    expect(restoreRequest(saved).rig).toBe('adult')
    expect(restoreRequest(null)).toBe(null)
    // Looks saved before sizes existed carry no rig: the server fits them.
    s.setItem('greyvar.appearance', JSON.stringify({ parts: { pants: 'skirt' }, palettes: {} }))
    expect(restoreRequest(loadSavedAppearance(s)).rig).toBe(undefined)
  })

  it('survives corrupt or unavailable storage', async () => {
    const { loadSavedAppearance, saveAppearance } = await import('./appearanceStore.js')
    expect(loadSavedAppearance({ getItem: () => '{nope' })).toBe(null)
    expect(loadSavedAppearance({ getItem: () => { throw new Error('blocked') } })).toBe(null)
    expect(() => saveAppearance({}, { setItem: () => { throw new Error('quota') } })).not.toThrow()
  })
})
