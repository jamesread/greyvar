import { describe, it, expect } from 'vitest'
import { parseRiftArgs } from './riftArgs.js'

describe('parseRiftArgs', () => {
  it('defaults to the next level with a random layout', () => {
    expect(parseRiftArgs([])).toEqual({ request: { level: 0, seed: 0, mechanics: [], leave: false }, errors: [] })
  })

  it('reads a level, seed and mechanics', () => {
    const { request, errors } = parseRiftArgs(['5', 'seed=7', 'mechanics=lever,keyChest'])
    expect(errors).toEqual([])
    expect(request).toEqual({ level: 5, seed: 7, mechanics: ['lever', 'keyChest'], leave: false })
  })

  it('reads leave', () => {
    expect(parseRiftArgs(['leave']).request.leave).toBe(true)
    expect(parseRiftArgs(['leave', '3']).errors).toHaveLength(1)
  })

  it('rejects nonsense', () => {
    expect(parseRiftArgs(['abc']).errors).toHaveLength(1)
    expect(parseRiftArgs(['0']).errors).toHaveLength(1)
    expect(parseRiftArgs(['seed=-1']).errors).toHaveLength(1)
    expect(parseRiftArgs(['mechanics=lever,teleport']).errors[0]).toMatch(/teleport/)
    expect(parseRiftArgs(['colour=red']).errors).toHaveLength(1)
  })
})
