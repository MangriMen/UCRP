import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useAuditLogs } from './use-audit-logs'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('useAuditLogs', () => {
  it('filters entries by action and period', async () => {
    const byAction = renderHook(() => useAuditLogs({ action: 'user.update' }), { wrapper })
    await waitFor(() => expect(byAction.result.current.data).toBeDefined())
    expect(byAction.result.current.data?.items.map((l) => l.id)).toEqual(['audit-0003'])

    const byPeriod = renderHook(() => useAuditLogs({ from: '2026-10-29T00:00:00Z' }), { wrapper })
    await waitFor(() => expect(byPeriod.result.current.data).toBeDefined())
    expect(byPeriod.result.current.data?.items).toHaveLength(2)
  })
})
