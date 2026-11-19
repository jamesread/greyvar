// Per-browser memory of the player's chosen look. There are no accounts yet
// (each connection is a new playerN), so the client keeps the look and asks
// the server to re-apply it after joining; the server still validates it.

const KEY = 'greyvar.appearance'

/** Saved appearance request, or null. */
export function loadSavedAppearance (storage = globalThis.localStorage) {
  try {
    const raw = storage?.getItem(KEY)
    if (!raw) {
      return null
    }
    const a = JSON.parse(raw)
    if (a == null || typeof a !== 'object') {
      return null
    }
    return {
      parts: typeof a.parts === 'object' && a.parts ? a.parts : {},
      palettes: typeof a.palettes === 'object' && a.palettes ? a.palettes : {},
      heldItem: typeof a.heldItem === 'string' ? a.heldItem : ''
    }
  } catch {
    return null
  }
}

/** Remember an appearance (as confirmed by the server). */
export function saveAppearance (appearance, storage = globalThis.localStorage) {
  try {
    storage?.setItem(KEY, JSON.stringify({
      parts: appearance?.parts ?? {},
      palettes: appearance?.palettes ?? {},
      heldItem: appearance?.heldItem ?? ''
    }))
  } catch {
    // Private mode / blocked storage: the look just will not persist.
  }
}

/** Request that restores a saved look ("-" clears any held item). */
export function restoreRequest (saved) {
  if (saved == null) {
    return null
  }
  return { parts: saved.parts, palettes: saved.palettes, heldItem: saved.heldItem || '-' }
}
