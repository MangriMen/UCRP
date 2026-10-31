import { act, renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { applyGuestFilters, parseGuestFilters, useGuestFilters } from './use-guest-filters'

const wrapperFor = (url: string) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return <MemoryRouter initialEntries={[url]}>{children}</MemoryRouter>
  }

describe('parseGuestFilters', () => {
  it('returns empty filters and the first page by default', () => {
    expect(parseGuestFilters(new URLSearchParams())).toEqual({
      q: '',
      sourceId: '',
      tenantId: '',
      periodFrom: '',
      periodTo: '',
      page: 1,
    })
  })

  it('reads snake_case params and drops invalid dates and pages', () => {
    const params = new URLSearchParams(
      'q= Иван &source_id=ds-1&tenant_id=t-1&period_from=2026-01-05&period_to=bad&page=-2',
    )
    expect(parseGuestFilters(params)).toEqual({
      q: 'Иван',
      sourceId: 'ds-1',
      tenantId: 't-1',
      periodFrom: '2026-01-05',
      periodTo: '',
      page: 1,
    })
  })
})

describe('applyGuestFilters', () => {
  it('sets and clears params and resets the page', () => {
    const next = applyGuestFilters(new URLSearchParams('tenant_id=t-1&page=3'), {
      tenantId: '',
      sourceId: 'ds-2',
    })
    expect(next.toString()).toBe('source_id=ds-2')
  })
})

describe('useGuestFilters', () => {
  it('updates, paginates and resets through the URL', () => {
    const { result } = renderHook(() => useGuestFilters(), {
      wrapper: wrapperFor('/guests?q=ab'),
    })
    expect(result.current.isFiltered).toBe(true)

    act(() => result.current.update({ tenantId: 't-9' }))
    expect(result.current.filters.tenantId).toBe('t-9')

    act(() => result.current.setPage(2))
    expect(result.current.filters.page).toBe(2)

    act(() => result.current.reset())
    expect(result.current.isFiltered).toBe(false)
    expect(result.current.filters.page).toBe(1)
  })
})
