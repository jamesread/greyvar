// The map cell a character faces: a point 12px from its 16px cell centre in
// the facing direction (diagonals reach both ways), so the target is the
// neighbouring cell the player is looking at.

const REACH_PX = 12

export function facedCell (x, y, facing = 'Down', tileSize = 16) {
  const f = String(facing ?? 'Down')
  const dx = f.includes('Left') ? -1 : f.includes('Right') ? 1 : 0
  const dy = f.includes('Up') ? -1 : f.includes('Down') ? 1 : 0
  const cx = x + tileSize / 2 + dx * REACH_PX
  const cy = y + tileSize / 2 + dy * REACH_PX
  return { row: Math.max(0, Math.floor(cy / tileSize)), col: Math.max(0, Math.floor(cx / tileSize)) }
}
