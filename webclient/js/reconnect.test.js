import { describe, it, expect } from 'vitest'
import { reconnectDelayMs } from './reconnect.js'

describe('reconnectDelayMs', () => {
  it('backs off exponentially and caps at 5s', () => {
    expect([0, 1, 2, 3, 4, 10].map(reconnectDelayMs)).toEqual([500, 1000, 2000, 4000, 5000, 5000])
  })
})
