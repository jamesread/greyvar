// Where a world sprite is actually drawn. Most sprites are drawn from their
// position (top-left origin); characters (LayeredSprite) are drawn at an
// `offset` from their cell so their feet land on it, and an adult rises above
// the cell. Use this, not img.x/img.y, to place things over or under one.

/** World bounds {left, top, right, bottom} of a sprite as drawn. */
export function spriteBounds (img) {
  const off = img?.offset ?? { x: 0, y: 0 }
  const w = img?.displayWidth || 16
  const h = img?.displayHeight || 16
  const left = (img?.x ?? 0) + off.x
  const top = (img?.y ?? 0) + off.y
  return { left, top, right: left + w, bottom: top + h }
}
