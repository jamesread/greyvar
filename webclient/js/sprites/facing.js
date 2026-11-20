// Which way a character faces, derived from its last walk/bump/use/face state.

const DIRECTIONAL = /^(?:walk|bump|use|face)(Up|Down)?(Left|Right)?$/

/**
 * "walkUpLeft" -> "UpLeft", "bumpRight" -> "Right"; null for non-directional
 * states (idle, pressed, ...).
 */
export function facingFromState (state) {
  const m = DIRECTIONAL.exec(String(state))
  if (m == null || (m[1] == null && m[2] == null)) {
    return null
  }
  return (m[1] ?? '') + (m[2] ?? '')
}

/**
 * Idle clip for a facing: idle<Facing>, then the horizontal part for
 * diagonals (diagonal walks use the side view), else null meaning "use the
 * default (front) idle with its random variants".
 */
export function idleClipForFacing (clips, facing) {
  if (facing == null || facing === 'Down') {
    return null
  }
  const candidates = [`idle${facing}`]
  const horizontal = /Left|Right/.exec(facing)?.[0]
  if (horizontal != null) {
    candidates.push(`idle${horizontal}`)
  }
  return candidates.find((name) => Array.isArray(clips[name]) && clips[name].length > 0) ?? null
}

/**
 * Use clip for a facing: use<Facing>, then the side clip for diagonals
 * (useUpLeft -> useLeft), then useDown. Null if the rig has no use clips.
 */
export function useClipForFacing (clips, facing) {
  const f = String(facing ?? 'Down')
  const horizontal = /Left|Right/.exec(f)?.[0]
  const candidates = [`use${f}`, horizontal != null ? `use${horizontal}` : null, 'useDown']
  return candidates.find((name) => name != null && Array.isArray(clips[name]) && clips[name].length > 0) ?? null
}
