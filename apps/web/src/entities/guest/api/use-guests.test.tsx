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
  it('lists all guests without filters', async () => {
    const { result } = renderHook(() => useGuestSearch({}), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((g) => g.id)).toEqual([
      'guest-0001',
      'guest-0002',
      'guest-0003',
    ])
  })

  it('ignores a query shorter than the minimum', async () => {
    const { result } = renderHook(() => useGuestSearch({ query: 'и' }), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items).toHaveLength(3)
  })

  it('filters by tenant, source and period', async () => {
    const { result } = renderHook(
      () =>
        useGuestSearch({
          tenantId: 'tenant-0001',
          sourceId: 'ds-0001',
          periodFrom: '2026-10-01',
          periodTo: '2026-10-02',
        }),
      { wrapper },
    )
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((g) => g.id)).toEqual(['guest-0002'])
  })

  it('finds guests by name and excludes merged profiles', async () => {
    const { result } = renderHook(() => useGuestSearch({ query: 'Петр' }), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((g) => g.id)).toEqual(['guest-0001'])
  })
})
