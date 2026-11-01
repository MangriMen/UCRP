import { describe, expect, it } from 'vitest'
import { formatMoney, formatSuppressed } from './format'

describe('formatMoney', () => {
  it('converts minor units to rubles', () => {
    expect(formatMoney(35000).replace(/\s/g, ' ')).toBe('350,00 ₽')
  })
})

describe('formatSuppressed', () => {
  it('returns the suppressed label for a missing value', () => {
    expect(formatSuppressed(null, String)).toBe('скрыто')
    expect(formatSuppressed(undefined, String)).toBe('скрыто')
  })

  it('formats a present value, including zero', () => {
    expect(formatSuppressed(0, (value) => `#${value}`)).toBe('#0')
  })
})
