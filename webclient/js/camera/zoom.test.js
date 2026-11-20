import { describe, it, expect } from 'vitest'
import { parseZoomArg, loadZoom, saveZoom, DEFAULT_ZOOM, MAX_ZOOM, MIN_ZOOM } from './zoom.js'

const mem = () => {
  const m = new Map()
  return { getItem: (k) => m.get(k) ?? null, setItem: (k, v) => m.set(k, v) }
}

describe('camera zoom', () => {
  it('parses absolute, relative and named zooms', () => {
    expect(parseZoomArg('4', 6)).toEqual({ zoom: 4 })
    expect(parseZoomArg('in', 6)).toEqual({ zoom: 7 })
    expect(parseZoomArg('-', 6)).toEqual({ zoom: 5 })
    expect(parseZoomArg('+3', 6)).toEqual({ zoom: 9 })
    expect(parseZoomArg('-2', 6)).toEqual({ zoom: 4 })
    expect(parseZoomArg('reset', 3)).toEqual({ zoom: DEFAULT_ZOOM })
  })

  it('clamps to whole numbers in range and rejects junk', () => {
    expect(parseZoomArg('99', 6)).toEqual({ zoom: MAX_ZOOM })
    expect(parseZoomArg('0', 6)).toEqual({ zoom: MIN_ZOOM })
    expect(parseZoomArg('out', MIN_ZOOM)).toEqual({ zoom: MIN_ZOOM })
    expect(parseZoomArg('2.5', 6).error).toMatch(/bad zoom/)
    expect(parseZoomArg('big', 6).error).toMatch(/bad zoom/)
  })

  it('remembers the zoom and survives bad storage', () => {
    const s = mem()
    expect(loadZoom(s)).toBe(DEFAULT_ZOOM)
    saveZoom(3, s)
    expect(loadZoom(s)).toBe(3)
    s.setItem('greyvar.zoom', 'nope')
    expect(loadZoom(s)).toBe(DEFAULT_ZOOM)
    expect(loadZoom({ getItem: () => { throw new Error('blocked') } })).toBe(DEFAULT_ZOOM)
  })
})
