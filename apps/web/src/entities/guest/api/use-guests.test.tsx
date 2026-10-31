import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useGuestSearch } from './use-guests'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('useGuestSearch', () => {
  it('does not fetch a query shorter than the minimum', () => {
    const { result } = renderHook(() => useGuestSearch({ query: 'и' }), { wrapper })
    expect(result.current.fetchStatus).toBe('idle')
  })

  it('finds guests by name and excludes merged profiles', async () => {
    const { result } = renderHook(() => useGuestSearch({ query: 'Петр' }), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((g) => g.id)).toEqual(['guest-0001'])
  })
})
