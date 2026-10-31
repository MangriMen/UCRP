import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { mockAggregate } from '../mocks/handlers'
import { useTenantAggregates } from './use-tenant-aggregates'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('useTenantAggregates', () => {
  it('returns a suppressed count as null', async () => {
    const { result } = renderHook(() => useTenantAggregates('tenant-0001'), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data).toEqual(mockAggregate)
  })

  it('does not fetch without a tenant', () => {
    const { result } = renderHook(() => useTenantAggregates(null), { wrapper })
    expect(result.current.fetchStatus).toBe('idle')
  })

  it('requests the daily series for a period', async () => {
    const period = { from: '2026-10-01', to: '2026-10-03' }
    const { result } = renderHook(() => useTenantAggregates('tenant-0002', period), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    const data = result.current.data
    expect(data && 'series' in data ? data.series.map((day) => day.date) : []).toEqual([
      '2026-10-01',
      '2026-10-02',
      '2026-10-03',
    ])
  })

  it('surfaces forbidden_tenant for a foreign tenant', async () => {
    const { result } = renderHook(() => useTenantAggregates('tenant-0003'), { wrapper })
    await waitFor(() => expect(result.current.error).not.toBeNull())
  })
})
