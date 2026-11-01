import { describe, expect, it } from 'vitest'
import { guestDisplayName } from './labels'

describe('guestDisplayName', () => {
  it('returns the name when present', () => {
    expect(guestDisplayName('Иван Петров')).toBe('Иван Петров')
  })

  it.each([null, undefined, '', '   '])('falls back to the unnamed label for %j', (value) => {
    expect(guestDisplayName(value)).toBe('Без имени')
  })
})
