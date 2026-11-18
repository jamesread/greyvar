import { describe, expect, it } from 'vitest'
import { textureUrl } from './client.js'

describe('textureUrl', () => {
  it('builds a res path for tile textures', () => {
    expect(textureUrl('tiles', 'water.png')).toBe('/res/img/textures/tiles/water.png')
  })
})
