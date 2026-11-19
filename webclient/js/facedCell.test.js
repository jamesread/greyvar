import { describe, it, expect } from 'vitest'
import { facedCell } from './facedCell.js'

describe('facedCell', () => {
  it('targets the neighbouring cell in the facing direction', () => {
    expect(facedCell(16, 16, 'Down')).toEqual({ row: 2, col: 1 })
    expect(facedCell(16, 16, 'Up')).toEqual({ row: 0, col: 1 })
    expect(facedCell(16, 16, 'Left')).toEqual({ row: 1, col: 0 })
    expect(facedCell(16, 16, 'Right')).toEqual({ row: 1, col: 2 })
    expect(facedCell(16, 16, 'UpRight')).toEqual({ row: 0, col: 2 })
  })

  it('works mid-cell and never goes negative', () => {
    expect(facedCell(20, 16, 'Right')).toEqual({ row: 1, col: 2 })
    expect(facedCell(0, 0, 'UpLeft')).toEqual({ row: 0, col: 0 })
    expect(facedCell(16, 16, undefined)).toEqual({ row: 2, col: 1 })
  })
})
