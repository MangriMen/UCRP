import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useTenantAggregates } from './use-tenant-aggregates'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

const period = { from: '2026-05-01T00:00:00Z', to: '2026-10-31T23:59:59Z' }

describe('useTenantAggregates', () => {
  it('returns suppressed cells as null', async () => {
    const { result } = renderHook(() => useTenantAggregates('tenant-0001', period), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.suppressed).toBe(true)
    expect(result.current.data?.series[1].newGuests).toBeNull()
  })

  it('does not fetch without a tenant', () => {
    const { result } = renderHook(() => useTenantAggregates(null, period), { wrapper })
    expect(result.current.fetchStatus).toBe('idle')
  })

  it('surfaces forbidden_tenant for a foreign tenant', async () => {
    const { result } = renderHook(() => useTenantAggregates('tenant-0002', period), { wrapper })
    await waitFor(() => expect(result.current.error).not.toBeNull())
  })
})
