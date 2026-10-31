import type { Query } from '@tanstack/react-query'
import { describe, expect, it } from 'vitest'
import { pollWhile } from './polling'

const queryWith = (data?: string) => ({ state: { data } }) as Query<string, Error, string>

describe('pollWhile', () => {
  const refetchInterval = pollWhile((status: string) => status === 'processing', 2000)

  it('keeps polling while the status is active', () => {
    expect(refetchInterval(queryWith('processing'))).toBe(2000)
  })

  it('stops on a terminal status and before first data', () => {
    expect(refetchInterval(queryWith('done'))).toBe(false)
    expect(refetchInterval(queryWith())).toBe(false)
  })
})
