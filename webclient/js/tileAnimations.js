// Tiled tile animations (tileset <animation>), sent with each grid's tilesets.
// Pure helpers; GridScene turns these into Phaser animations.

export function tileAnimKey (atlasKey, tileId) {
  return `tile:${atlasKey}:${tileId}`
}

/** Normalise a tileset's animations from a grid message. */
export function tilesetAnimationsFromMessage (tileset) {
  return (tileset.animations ?? [])
    .map((a) => ({
      tileId: a.tileId ?? 0,
      frames: a.frames ?? [],
      durationsMs: a.durationsMs ?? [],
      synced: a.synced === true
    }))
    .filter((a) => a.frames.length >= 2)
}

/** Add a grid's tileset animations to `index` (Map animKey -> animation). */
export function indexTileAnimations (tilesets, index = new Map()) {
  for (const tileset of tilesets ?? []) {
    for (const anim of tileset.animations ?? []) {
      index.set(tileAnimKey(tileset.key, anim.tileId), { atlasKey: tileset.key, ...anim })
    }
  }
  return index
}

/** Phaser animation config: per-frame durations, looping forever. */
export function phaserTileAnimConfig (anim) {
  return {
    key: tileAnimKey(anim.atlasKey, anim.tileId),
    frames: anim.frames.map((frame, i) => ({
      key: anim.atlasKey,
      frame,
      duration: anim.durationsMs[i] > 0 ? anim.durationsMs[i] : 100
    })),
    repeat: -1
  }
}

/**
 * Start frame for a synced animation: where a shared clock (ms) is in the
 * cycle, so every cell, including ones drawn later (a map scrolling in), is
 * on the same frame and patterns flow across tiles.
 */
export function syncedStartFrame (anim, nowMs) {
  const durations = anim.frames.map((_, i) => (anim.durationsMs[i] > 0 ? anim.durationsMs[i] : 100))
  const cycle = durations.reduce((a, b) => a + b, 0)
  let t = ((nowMs % cycle) + cycle) % cycle
  for (let i = 0; i < durations.length; i++) {
    if (t < durations[i]) {
      return i
    }
    t -= durations[i]
  }
  return 0
}

/**
 * Deterministic start frame per cell so a field of the same animated tile
 * does not move in lockstep, yet looks identical on every client and reload.
 */
export function tileAnimStartFrame (row, col, frameCount) {
  if (frameCount <= 1) {
    return 0
  }
  // Integer hash of the cell (Knuth multiplicative mixing).
  let h = (Math.imul(row + 1, 73856093) ^ Math.imul(col + 1, 19349663)) >>> 0
  h = Math.imul(h ^ (h >>> 15), 2246822519) >>> 0
  return h % frameCount
}
