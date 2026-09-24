import { api, textureUrl } from './api.js'

const app = document.getElementById('app')

const state = {
  tab: 'entdefs',
  health: null,
  entdefs: [],
  tiledefs: [],
  entityTextures: [],
  tileTextures: [],
  selectedEntdef: null,
  selectedTiledef: null,
  entdefDraft: null,
  tiledefDraft: null,
  status: '',
  error: ''
}

function el (tag, attrs = {}, children = []) {
  const node = document.createElement(tag)
  for (const [key, value] of Object.entries(attrs)) {
    if (key === 'className') node.className = value
    else if (key === 'text') node.textContent = value
    else if (key.startsWith('on')) node.addEventListener(key.slice(2).toLowerCase(), value)
    else node.setAttribute(key, value)
  }
  for (const child of children) {
    if (child == null) continue
    node.append(typeof child === 'string' ? document.createTextNode(child) : child)
  }
  return node
}

function setStatus (message) {
  state.status = message
  state.error = ''
  render()
}

function setError (message) {
  state.error = message
  state.status = ''
  render()
}

function defaultEntdefDraft () {
  return {
    title: '',
    initialState: 'idle',
    solid: false,
    states: {
      idle: { name: 'idle' }
    }
  }
}

function defaultTiledefDraft () {
  return {
    traversable: true,
    texture: ''
  }
}

function clone (value) {
  return JSON.parse(JSON.stringify(value))
}

async function loadAll () {
  try {
    const [health, entdefs, tiledefs, entityTextures, tileTextures] = await Promise.all([
      api.health(),
      api.listEntdefs(),
      api.listTiledefs(),
      api.listEntityTextures(),
      api.listTileTextures()
    ])
    state.health = health
    state.entdefs = Array.isArray(entdefs) ? entdefs : (entdefs.items || [])
    state.tiledefs = Array.isArray(tiledefs) ? tiledefs : (tiledefs.items || [])
    state.entityTextures = Array.isArray(entityTextures) ? entityTextures : (entityTextures.items || [])
    state.tileTextures = Array.isArray(tileTextures) ? tileTextures : (tileTextures.items || [])
    state.error = ''
    render()
  } catch (err) {
    setError(err.message)
  }
}

async function selectEntdef (name) {
  try {
    const item = await api.getEntdef(name)
    state.selectedEntdef = name
    state.entdefDraft = clone(item)
    setStatus(`Loaded entdef ${name}`)
  } catch (err) {
    setError(err.message)
  }
}

async function selectTiledef (name) {
  try {
    const item = await api.getTiledef(name)
    state.selectedTiledef = name
    state.tiledefDraft = clone(item)
    setStatus(`Loaded tiledef ${name}`)
  } catch (err) {
    setError(err.message)
  }
}

function renderStatesEditor (draft) {
  const rows = Object.entries(draft.states || {}).map(([key, value]) => {
    const nameInput = el('input', {
      value: value.name || key,
      oninput: (e) => {
        draft.states[key].name = e.target.value
      }
    })

    const removeBtn = el('button', {
      className: 'btn danger',
      text: 'Remove',
      onclick: () => {
        delete draft.states[key]
        render()
      }
    })

    return el('tr', {}, [
      el('td', { text: key }),
      el('td', {}, [nameInput]),
      el('td', { className: 'hint', text: `tileset type: ${(draft.title || '…')}_${key} or ${draft.title || '…'}` }),
      el('td', {}, [removeBtn])
    ])
  })

  const addStateBtn = el('button', {
    className: 'btn',
    text: 'Add state',
    onclick: () => {
      const name = prompt('State key (e.g. idle):')
      if (!name) return
      draft.states[name] = { name }
      render()
    }
  })

  return el('div', { className: 'form-grid' }, [
    el('p', {
      className: 'hint',
      text: 'Visuals come from tileset tile type/class matching the entdef title (and title_state for non-initial states).'
    }),
    el('table', { className: 'states-table' }, [
      el('thead', {}, [
        el('tr', {}, [
          el('th', { text: 'Key' }),
          el('th', { text: 'Display name' }),
          el('th', { text: 'Tileset match' }),
          el('th', { text: '' })
        ])
      ]),
      el('tbody', {}, rows)
    ]),
    addStateBtn
  ])
}

function renderEntdefEditor () {
  if (!state.entdefDraft) {
    return el('div', { className: 'empty', text: 'Select an entity definition or create a new one.' })
  }

  const draft = state.entdefDraft

  const initialStateSelect = el('select', {
    onchange: (e) => {
      draft.initialState = e.target.value
    }
  }, Object.keys(draft.states || {}).map((key) => el('option', {
    value: key,
    text: key,
    ...(draft.initialState === key ? { selected: 'selected' } : {})
  })))

  return el('div', { className: 'form-grid' }, [
    el('label', {}, ['Title (must match tileset type/class)', el('input', {
      value: draft.title || '',
      oninput: (e) => { draft.title = e.target.value; render() }
    })]),
    el('label', {}, [
      'Solid',
      el('select', {
        onchange: (e) => { draft.solid = e.target.value === 'true' }
      }, [
        el('option', { value: 'false', text: 'false', ...(!draft.solid ? { selected: 'selected' } : {}) }),
        el('option', { value: 'true', text: 'true', ...(draft.solid ? { selected: 'selected' } : {}) })
      ])
    ]),
    el('label', {}, ['Initial state', initialStateSelect]),
    el('h3', { text: 'States' }),
    renderStatesEditor(draft),
    el('div', { className: 'actions' }, [
      el('button', {
        className: 'btn primary',
        text: 'Save',
        onclick: async () => {
          try {
            await api.saveEntdef(state.selectedEntdef, draft)
            await loadAll()
            setStatus(`Saved ${state.selectedEntdef}`)
          } catch (err) {
            setError(err.message)
          }
        }
      }),
      el('button', {
        className: 'btn danger',
        text: 'Delete',
        onclick: async () => {
          if (!confirm(`Delete entdef ${state.selectedEntdef}?`)) return
          try {
            await api.deleteEntdef(state.selectedEntdef)
            state.selectedEntdef = null
            state.entdefDraft = null
            await loadAll()
            setStatus('Deleted entdef')
          } catch (err) {
            setError(err.message)
          }
        }
      })
    ])
  ])
}

function renderTiledefEditor () {
  if (!state.tiledefDraft) {
    return el('div', { className: 'empty', text: 'Select a tile definition or create a new one.' })
  }

  const draft = state.tiledefDraft
  const texturePreview = draft.texture
    ? el('img', { src: textureUrl('tiles', draft.texture), alt: draft.texture })
    : el('div', { className: 'empty', text: 'No texture selected' })

  const textureSelect = el('select', {
    onchange: (e) => {
      draft.texture = e.target.value
      render()
    }
  }, [
    el('option', { value: '', text: 'Select texture…' }),
    ...state.tileTextures.map((name) => el('option', {
      value: name,
      text: name,
      ...(draft.texture === name ? { selected: 'selected' } : {})
    }))
  ])

  return el('div', { className: 'form-grid' }, [
    el('label', {}, ['Texture filename', textureSelect]),
    el('div', { className: 'preview-row' }, [texturePreview]),
    el('label', {}, [
      'Traversable',
      el('select', {
        onchange: (e) => { draft.traversable = e.target.value === 'true' }
      }, [
        el('option', { value: 'true', text: 'true', ...(draft.traversable ? { selected: 'selected' } : {}) }),
        el('option', { value: 'false', text: 'false', ...(!draft.traversable ? { selected: 'selected' } : {}) })
      ])
    ]),
    el('div', { className: 'actions' }, [
      el('button', {
        className: 'btn primary',
        text: 'Save',
        onclick: async () => {
          try {
            await api.saveTiledef(state.selectedTiledef, draft)
            await loadAll()
            setStatus(`Saved ${state.selectedTiledef}`)
          } catch (err) {
            setError(err.message)
          }
        }
      }),
      el('button', {
        className: 'btn danger',
        text: 'Delete',
        onclick: async () => {
          if (!confirm(`Delete tiledef ${state.selectedTiledef}?`)) return
          try {
            await api.deleteTiledef(state.selectedTiledef)
            state.selectedTiledef = null
            state.tiledefDraft = null
            await loadAll()
            setStatus('Deleted tiledef')
          } catch (err) {
            setError(err.message)
          }
        }
      })
    ])
  ])
}

function renderSidebar () {
  const isEnt = state.tab === 'entdefs'
  const items = isEnt ? state.entdefs : state.tiledefs
  const selected = isEnt ? state.selectedEntdef : state.selectedTiledef

  const newBtn = el('button', {
    className: 'btn',
    text: 'New',
    onclick: () => {
      const name = prompt(`New ${isEnt ? 'entdef' : 'tiledef'} name:`)
      if (!name) return
      ;(async () => {
        try {
          if (isEnt) {
            const draft = defaultEntdefDraft()
            await api.createEntdef(name, draft)
            await loadAll()
            await selectEntdef(name)
            setStatus(`Created entdef ${name}`)
          } else {
            const draft = defaultTiledefDraft()
            await api.createTiledef(name, draft)
            await loadAll()
            await selectTiledef(name)
            setStatus(`Created tiledef ${name}`)
          }
        } catch (err) {
          setError(err.message)
        }
      })()
    }
  })

  const list = el('ul', { className: 'list' }, items.map((item) => {
    const name = item.name
    const subtitle = isEnt
      ? `${item.title || name}${item.solid ? ' · solid' : ''} · ${item.stateCount || 0} states`
      : `${item.traversable ? 'walkable' : 'blocked'}`
    return el('li', {}, [
      el('button', {
        className: `list-item${selected === name ? ' active' : ''}`,
        onclick: () => (isEnt ? selectEntdef(name) : selectTiledef(name))
      }, [
        el('span', { text: name }),
        el('small', { text: subtitle })
      ])
    ])
  }))

  return el('aside', { className: 'sidebar' }, [
    el('div', { className: 'sidebar-header' }, [
      el('strong', { text: isEnt ? 'Entity definitions' : 'Tile definitions' }),
      newBtn
    ]),
    list
  ])
}

function renderPanel () {
  const title = state.tab === 'entdefs'
    ? (state.selectedEntdef || 'Entity editor')
    : (state.selectedTiledef || 'Tile editor')

  return el('section', { className: 'editor-panel' }, [
    el('div', { className: 'panel-header' }, [
      el('strong', { text: title })
    ]),
    el('div', { className: 'panel-body' }, [
      state.tab === 'entdefs' ? renderEntdefEditor() : renderTiledefEditor(),
      state.error ? el('div', { className: 'error', text: state.error }) : null,
      el('div', { className: 'status', text: state.status })
    ])
  ])
}

function render () {
  const healthText = state.health
    ? `dat: ${state.health.datDir} · res: ${state.health.resDir}`
    : 'Connecting…'

  app.replaceChildren(el('div', { className: 'app-shell' }, [
    el('header', { className: 'topbar' }, [
      el('div', {}, [
        el('h1', { text: 'Greyvar Dat Editor' }),
        el('div', { className: 'meta', text: healthText })
      ]),
      el('div', { className: 'tabs' }, [
        el('button', {
          className: `tab${state.tab === 'entdefs' ? ' active' : ''}`,
          text: 'Entdefs',
          onclick: () => { state.tab = 'entdefs'; render() }
        }),
        el('button', {
          className: `tab${state.tab === 'tiledefs' ? ' active' : ''}`,
          text: 'Tiledefs',
          onclick: () => { state.tab = 'tiledefs'; render() }
        })
      ])
    ]),
    el('main', { className: 'layout' }, [
      renderSidebar(),
      renderPanel()
    ])
  ]))
}

render()
loadAll()
