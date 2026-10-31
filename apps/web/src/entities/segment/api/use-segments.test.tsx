import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useSegmentMembers, useSegments, useSegmentVersions } from './use-segments'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('segment hooks', () => {
  it('filters segments by status', async () => {
    const { result } = renderHook(() => useSegments({ status: 'ready' }), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((s) => s.id)).toEqual([
      'segment-0001',
      'segment-0002',
      'segment-0005',
    ])
  })

  it('pages members and reports the total', async () => {
    const { result } = renderHook(() => useSegmentMembers('segment-0001', { page: 2 }), {
      wrapper,
    })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.meta.total).toBe(120)
    expect(result.current.data?.meta.page).toBe(2)
  })

  it('lists versions newest first', async () => {
    const { result } = renderHook(() => useSegmentVersions('segment-0001'), { wrapper })
    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.map((v) => v.versionNo)).toEqual([2, 1])
  })
})
