import { describe, expect, it } from 'vitest'
import { formatMoney } from './format'

describe('formatMoney', () => {
  it('converts minor units to rubles', () => {
    expect(formatMoney(35000).replace(/\s/g, ' ')).toBe('350,00 ₽')
  })
})
