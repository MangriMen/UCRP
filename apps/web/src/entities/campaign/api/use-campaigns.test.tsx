import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useCampaignResults, useCampaigns } from './use-campaigns'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('campaign hooks', () => {
  it('filters campaigns by status', async () => {
    const { result } = renderHook(() => useCampaigns({ status: 'finished' }), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((c) => c.id)).toEqual(['campaign-0002'])
  })

  it('loads results and fails for a campaign without them', async () => {
    const ok = renderHook(() => useCampaignResults('campaign-0002'), { wrapper })
    await waitFor(() => expect(ok.result.current.data).toBeDefined())
    expect(ok.result.current.data?.isSignificant).toBe(true)

    const missing = renderHook(() => useCampaignResults('campaign-0003'), { wrapper })
    await waitFor(() => expect(missing.result.current.error).not.toBeNull())
  })
})
