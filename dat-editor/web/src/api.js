const jsonHeaders = { 'Content-Type': 'application/json' }

async function request (path, options = {}) {
  const res = await fetch(path, options)
  const body = await res.json().catch(() => ({}))

  if (!res.ok) {
    throw new Error(body.error || `${res.status} ${res.statusText}`)
  }

  return body
}

export const api = {
  health: () => request('/api/health'),
  listEntdefs: () => request('/api/entdefs'),
  getEntdef: (name) => request(`/api/entdefs/${encodeURIComponent(name)}`),
  saveEntdef: (name, payload) => request(`/api/entdefs/${encodeURIComponent(name)}`, {
    method: 'PUT',
    headers: jsonHeaders,
    body: JSON.stringify(payload)
  }),
  createEntdef: (name, payload) => request('/api/entdefs', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ name, ...payload })
  }),
  deleteEntdef: (name) => request(`/api/entdefs/${encodeURIComponent(name)}`, { method: 'DELETE' }),
  listTiledefs: () => request('/api/tiledefs'),
  getTiledef: (name) => request(`/api/tiledefs/${encodeURIComponent(name)}`),
  saveTiledef: (name, payload) => request(`/api/tiledefs/${encodeURIComponent(name)}`, {
    method: 'PUT',
    headers: jsonHeaders,
    body: JSON.stringify(payload)
  }),
  createTiledef: (name, payload) => request('/api/tiledefs', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify({ name, ...payload })
  }),
  deleteTiledef: (name) => request(`/api/tiledefs/${encodeURIComponent(name)}`, { method: 'DELETE' }),
  listEntityTextures: () => request('/api/textures/entities'),
  listTileTextures: () => request('/api/textures/tiles')
}

export function textureUrl (kind, filename) {
  return `/res/img/textures/${kind}/${encodeURIComponent(filename)}`
}
