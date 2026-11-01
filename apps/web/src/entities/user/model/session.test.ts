import { QueryClient } from '@tanstack/react-query'
import { describe, expect, it } from 'vitest'
import { endSession, returnPathFor, startSession } from './session'
import type { Me } from './types'

const user = (userId: string) =>
  ({
    userId,
    email: `${userId}@x`,
    role: 'marketer',
    displayName: userId,
    permissions: [],
    policyVersion: 1,
  }) as Me

describe('returnPathFor', () => {
  it('keeps the return path for the same user and drops it for another one after logout', () => {
    const queryClient = new QueryClient()
    expect(returnPathFor(user('first'), '/segments/1')).toBe('/segments/1')

    startSession(queryClient, 'token', user('first'))
    endSession(queryClient)

    expect(returnPathFor(user('first'), '/segments/1')).toBe('/segments/1')
    expect(returnPathFor(user('second'), '/segments/1')).toBe('/')
  })
})
