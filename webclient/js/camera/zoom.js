// Camera zoom: how many screen pixels one art pixel takes. Whole numbers only,
// so pixel art stays crisp (no uneven pixel widths). Remembered per browser.

export const DEFAULT_ZOOM = 4
export const MIN_ZOOM = 1
export const MAX_ZOOM = 16
const KEY = 'greyvar.zoom'

export function clampZoom (z) {
  return Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, Math.round(z)))
}

/** The saved zoom, or the default. */
export function loadZoom (storage = globalThis.localStorage) {
  try {
    const z = Number(storage?.getItem(KEY))
    return Number.isFinite(z) && z > 0 ? clampZoom(z) : DEFAULT_ZOOM
  } catch {
    return DEFAULT_ZOOM
  }
}

export function saveZoom (z, storage = globalThis.localStorage) {
  try {
    storage?.setItem(KEY, String(z))
  } catch {
    // Blocked storage: the zoom just will not persist.
  }
}

/**
 * The zoom a console argument asks for, given the current one:
 * a number, "in"/"+", "out"/"-", "+N"/"-N", or "reset".
 * Returns {zoom} or {error}.
 */
export function parseZoomArg (arg, current) {
  const a = String(arg ?? '').trim().toLowerCase()
  let z
  if (a === 'in' || a === '+') {
    z = current + 1
  } else if (a === 'out' || a === '-') {
    z = current - 1
  } else if (a === 'reset' || a === 'default') {
    z = DEFAULT_ZOOM
  } else if (/^[+-]\d+$/.test(a)) {
    z = current + Number(a)
  } else if (/^\d+$/.test(a)) {
    z = Number(a)
  } else {
    return { error: `bad zoom "${arg}" (want ${MIN_ZOOM}-${MAX_ZOOM}, in, out, +N, -N or reset)` }
  }
  return { zoom: clampZoom(z) }
}
