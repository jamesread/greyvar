export function entityId (value) {
  if (value == null) {
    return null
  }

  return typeof value === 'string' ? parseInt(value, 10) : value
}

export function normalizeEntitySpawn (ent) {
  return {
    ...ent,
    entityId: entityId(ent.entityId),
    x: ent.x ?? 0,
    y: ent.y ?? 0
  }
}

export function normalizeEntityPosition (entpos) {
  return {
    ...entpos,
    entityId: entityId(entpos.entityId),
    x: entpos.x ?? 0,
    y: entpos.y ?? 0
  }
}

export function normalizeEntityStateChange (change) {
  return {
    ...change,
    entityId: entityId(change.entityId)
  }
}
