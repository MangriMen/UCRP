import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useGuestRecommendations, useSegmentRecommendations } from './use-recommendations'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('recommendation hooks', () => {
  it('loads guest recommendations', async () => {
    const { result } = renderHook(() => useGuestRecommendations('guest-0001'), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((item) => item.rank)).toEqual([1, 2, 3])
  })

  it('loads segment top tenants', async () => {
    const { result } = renderHook(() => useSegmentRecommendations('segment-0001', false), {
      wrapper,
    })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items[0].guestCount).toBe(64)
  })
})
