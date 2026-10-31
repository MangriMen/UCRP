import { describe, expect, it } from 'vitest'
import { defaultPeriod, isPeriodValid, periodQuery } from './period'

describe('period', () => {
  it('builds a default period ending today', () => {
    expect(defaultPeriod(6, new Date('2026-10-31T12:00:00Z'))).toEqual({
      from: '2026-04-30',
      to: '2026-10-31',
    })
  })

  it('rejects empty and inverted periods', () => {
    expect(isPeriodValid({ from: '2026-01-01', to: '2026-01-01' })).toBe(true)
    expect(isPeriodValid({ from: '2026-02-01', to: '2026-01-01' })).toBe(false)
    expect(isPeriodValid({ from: '', to: '2026-01-01' })).toBe(false)
  })

  it('expands days to whole-day UTC bounds', () => {
    expect(periodQuery({ from: '2026-01-01', to: '2026-01-31' })).toEqual({
      from: '2026-01-01T00:00:00Z',
      to: '2026-01-31T23:59:59Z',
    })
  })
})
