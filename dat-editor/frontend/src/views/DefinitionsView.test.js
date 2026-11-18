import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref, computed } from 'vue'
import DefinitionsView from './DefinitionsView.vue'

const listEntdefs = vi.fn()
const listTiledefs = vi.fn()
const listTextures = vi.fn()

vi.mock('../composables/client.js', () => ({
  client: {
    listEntdefs: (...args) => listEntdefs(...args),
    listTiledefs: (...args) => listTiledefs(...args),
    listTextures: (...args) => listTextures(...args),
    getEntdef: vi.fn(),
    getTiledef: vi.fn(),
    createEntdef: vi.fn(),
    createTiledef: vi.fn(),
    updateEntdef: vi.fn(),
    updateTiledef: vi.fn(),
    deleteEntdef: vi.fn(),
    deleteTiledef: vi.fn(),
  },
}))

describe('DefinitionsView', () => {
  beforeEach(() => {
    listEntdefs.mockResolvedValue({
      items: [{ name: 'cake', title: 'cake', initialState: 'cakey', solid: true, stateCount: 1 }],
    })
    listTiledefs.mockResolvedValue({
      items: [{ name: 'water', texture: 'water.png', traversable: true, hasFile: true }],
    })
    listTextures.mockResolvedValue({ filenames: ['water.png'] })
  })

  it('renders entdefs in the sidebar after load', async () => {
    const wrapper = mount(DefinitionsView, {
      props: { kind: 'entdefs' },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('cake')
    expect(wrapper.findAll('.list-item')).toHaveLength(1)
  })

  it('renders tiledefs in the sidebar after load', async () => {
    const wrapper = mount(DefinitionsView, {
      props: { kind: 'tiledefs' },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('water')
    expect(wrapper.findAll('.list-item')).toHaveLength(1)
  })
})
