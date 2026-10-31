import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it } from 'vitest'
import { useImport, useImports } from './use-imports'

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}

describe('useImport', () => {
  it('polls an active batch until it reaches a terminal status', async () => {
    const { result } = renderHook(() => useImport('batch-0004'), { wrapper })

    await waitFor(() => expect(result.current.data?.status).toBe('running'))
    await waitFor(() => expect(result.current.data?.status).toBe('completed'), { timeout: 8000 })
  }, 10000)

  it('does not poll a finished batch', async () => {
    const { result } = renderHook(() => useImport('batch-0001'), { wrapper })

    await waitFor(() => expect(result.current.data?.status).toBe('completed'))
    expect(result.current.isFetching).toBe(false)
  })
})

describe('useImports', () => {
  it('filters the list by data source', async () => {
    const { result } = renderHook(() => useImports({ dataSourceId: 'ds-0002' }), { wrapper })

    await waitFor(() => expect(result.current.data).toBeDefined())
    expect(result.current.data?.items.map((b) => b.id)).toEqual(['batch-0002'])
  })
})
