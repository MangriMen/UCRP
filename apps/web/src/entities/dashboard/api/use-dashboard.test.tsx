import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useDashboardMetrics, usePenetration } from './use-dashboard'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

const period = { from: '2026-05-01T00:00:00Z', to: '2026-10-31T23:59:59Z' }

describe('dashboard hooks', () => {
  it('loads metrics for the period', async () => {
    const { result } = renderHook(() => useDashboardMetrics(period), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.matchRate).toBe(0.74)
  })

  it('does not fetch when disabled', () => {
    const { result } = renderHook(() => usePenetration(period, false), { wrapper })
    expect(result.current.fetchStatus).toBe('idle')
  })

  it('surfaces a rejected period as an error', async () => {
    const { result } = renderHook(
      () => usePenetration({ from: '2020-01-01T00:00:00Z', to: '2026-10-31T00:00:00Z' }),
      { wrapper },
    )
    await waitFor(() => expect(result.current.error).not.toBeNull())
  })
})
