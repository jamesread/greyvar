// Pure helpers for layered sprite rigs (see res/sprites/README.md).
// No Phaser imports here so these can be unit tested.

/**
 * Fill an appearance with rig defaults and drop values the rig does not know,
 * so a bad or stale appearance from the server can never break rendering.
 *
 * @param {object} rig parsed <name>.rig.json
 * @param {{parts?: object, palettes?: object, heldItem?: string}} [appearance]
 * @returns {{parts: Object<string,string>, palettes: Object<string,string>, heldItem: string}}
 */
export function resolveAppearance (rig, appearance = {}) {
  const parts = {}
  for (const layer of rig.layers) {
    const want = appearance?.parts?.[layer.name]
    parts[layer.name] = layer.variants.includes(want) ? want : layer.default
  }

  const palettes = {}
  for (const slot of rig.slotOrder) {
    const want = appearance?.palettes?.[slot]
    const known = rig.palettes[slot] ?? {}
    palettes[slot] = Object.hasOwn(known, want) ? want : rig.defaults.palettes[slot]
  }

  return { parts, palettes, heldItem: appearance?.heldItem ?? '' }
}

export function hexToInt (hex) {
  return parseInt(String(hex).replace('#', ''), 16)
}

/**
 * Key colour -> replacement colour (both 0xRRGGBB) for the given slots.
 * Identity entries are omitted, so an empty map means "use the sheet as is".
 */
export function recolourMap (rig, palettes, slots = rig.slotOrder) {
  const map = new Map()
  for (const slot of slots) {
    const keys = rig.slots[slot] ?? []
    const ramp = rig.palettes[slot]?.[palettes[slot]]
    if (ramp == null) {
      continue
    }
    keys.forEach((key, i) => {
      const from = hexToInt(key)
      const to = hexToInt(ramp[i])
      if (from !== to) {
        map.set(from, to)
      }
    })
  }
  return map
}

/** Recolour RGBA pixel data in place. Returns the number of pixels changed. */
export function recolourPixels (data, map) {
  if (map.size === 0) {
    return 0
  }
  let changed = 0
  for (let i = 0; i < data.length; i += 4) {
    if (data[i + 3] === 0) {
      continue
    }
    const rgb = (data[i] << 16) | (data[i + 1] << 8) | data[i + 2]
    const to = map.get(rgb)
    if (to === undefined) {
      continue
    }
    data[i] = (to >> 16) & 0xff
    data[i + 1] = (to >> 8) & 0xff
    data[i + 2] = to & 0xff
    changed++
  }
  return changed
}

/**
 * Texture cache key for one layer sheet under a palette choice. Only slots the
 * layer actually uses are included, so e.g. hair colour changes do not create
 * new shirt textures.
 */
export function recolouredTextureKey (sheetPath, layerSlots, palettes) {
  const parts = layerSlots.map((slot) => `${slot}=${palettes[slot]}`)
  return parts.length === 0 ? sheetPath : `${sheetPath}|${parts.join(',')}`
}

/**
 * Parse console args like ["hair=blonde", "hair=long", "held=key"] into an
 * AppearanceRequest. Slot and layer names can overlap (hair), so a value that
 * names a palette is a palette, otherwise a part variant.
 *
 * @returns {{request: {parts: object, palettes: object, heldItem?: string}, errors: string[]}}
 */
export function parseAppearanceArgs (rig, args) {
  const request = { parts: {}, palettes: {} }
  const errors = []
  for (const arg of args) {
    const [key, value] = String(arg).split('=', 2)
    if (!key || value == null || value === '') {
      errors.push(`bad argument "${arg}" (want key=value)`)
      continue
    }
    if (key === 'held') {
      request.heldItem = value
      continue
    }
    if (rig.palettes[key] != null && Object.hasOwn(rig.palettes[key], value)) {
      request.palettes[key] = value
      continue
    }
    const layer = rig.layers.find((l) => l.name === key)
    if (layer != null && layer.variants.includes(value)) {
      request.parts[key] = value
      continue
    }
    const options = [
      ...Object.keys(rig.palettes[key] ?? {}),
      ...(layer?.variants.length > 1 ? layer.variants : [])
    ]
    errors.push(options.length > 0
      ? `unknown ${key} "${value}" (have: ${options.join(', ')})`
      : `unknown key "${key}"`)
  }
  return { request, errors }
}

/** Human-readable list of every customisation option in a rig. */
export function describeAppearanceOptions (rig) {
  const rows = []
  for (const slot of rig.slotOrder) {
    rows.push(`  ${slot} colour: ${Object.keys(rig.palettes[slot] ?? {}).sort().join(', ')}`)
  }
  for (const layer of rig.layers) {
    if (layer.variants.length > 1) {
      rows.push(`  ${layer.name} style: ${layer.variants.join(', ')}`)
    }
  }
  rows.push('  held: <item name> or - to clear')
  return rows
}
