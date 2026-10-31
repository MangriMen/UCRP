import { describe, expect, it } from 'vitest'
import { cn } from './utils'

describe('cn', () => {
  it('keeps the last conflicting tailwind class', () => {
    expect(cn('p-2', 'p-4')).toBe('p-4')
  })
})
