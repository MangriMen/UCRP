import { describe, expect, it } from 'vitest'
import { validateLimit } from './validate-limit'

describe('validateLimit', () => {
  it('accepts values inside the contract ranges', () => {
    expect(validateLimit('100', '365')).toEqual({ maxContacts: undefined, periodDays: undefined })
  })

  it('rejects empty, fractional and out-of-range values', () => {
    expect(validateLimit('', '1').maxContacts).toBeDefined()
    expect(validateLimit('101', '1').maxContacts).toBeDefined()
    expect(validateLimit('1', '1.5').periodDays).toBeDefined()
    expect(validateLimit('1', '366').periodDays).toBeDefined()
  })
})
