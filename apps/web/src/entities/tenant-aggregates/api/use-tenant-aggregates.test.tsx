import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useTenantAggregates } from './use-tenant-aggregates'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('useTenantAggregates', () => {
  it('returns a suppressed count as null', async () => {
    const { result } = renderHook(() => useTenantAggregates('tenant-0001'), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data).toEqual({ visitsCount: null, suppressed: true })
  })

  it('does not fetch without a tenant', () => {
    const { result } = renderHook(() => useTenantAggregates(null), { wrapper })
    expect(result.current.fetchStatus).toBe('idle')
  })

  it('surfaces forbidden_tenant for a foreign tenant', async () => {
    const { result } = renderHook(() => useTenantAggregates('tenant-0002'), { wrapper })
    await waitFor(() => expect(result.current.error).not.toBeNull())
  })
})
