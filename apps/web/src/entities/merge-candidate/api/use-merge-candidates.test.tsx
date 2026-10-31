import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useMergeCandidate, useMergeCandidates } from './use-merge-candidates'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('useMergeCandidates', () => {
  it('filters the queue by resolution', async () => {
    const { result } = renderHook(() => useMergeCandidates({ resolution: 'pending' }), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((c) => c.id)).toEqual(['candidate-0001'])
  })

  it('loads both profiles of a candidate', async () => {
    const { result } = renderHook(() => useMergeCandidate('candidate-0001'), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.guestA.id).toBe('guest-0001')
    expect(result.current.data?.guestB.id).toBe('guest-0002')
  })
})
