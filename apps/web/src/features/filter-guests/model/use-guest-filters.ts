import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router'

export interface GuestFilters {
  q: string
  sourceId: string
  tenantId: string
  periodFrom: string
  periodTo: string
  page: number
}

const PARAM_NAMES = {
  q: 'q',
  sourceId: 'source_id',
  tenantId: 'tenant_id',
  periodFrom: 'period_from',
  periodTo: 'period_to',
} as const

type FilterKey = keyof typeof PARAM_NAMES

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/

const parseDate = (value: string | null) => (value && DATE_PATTERN.test(value) ? value : '')

const parsePage = (value: string | null) => {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

export function parseGuestFilters(params: URLSearchParams): GuestFilters {
  return {
    q: params.get(PARAM_NAMES.q)?.trim() ?? '',
    sourceId: params.get(PARAM_NAMES.sourceId) ?? '',
    tenantId: params.get(PARAM_NAMES.tenantId) ?? '',
    periodFrom: parseDate(params.get(PARAM_NAMES.periodFrom)),
    periodTo: parseDate(params.get(PARAM_NAMES.periodTo)),
    page: parsePage(params.get('page')),
  }
}

export function applyGuestFilters(
  current: URLSearchParams,
  patch: Partial<Record<FilterKey, string>>,
): URLSearchParams {
  const next = new URLSearchParams(current)
  for (const [key, value] of Object.entries(patch)) {
    const name = PARAM_NAMES[key as FilterKey]
    if (value) next.set(name, value)
    else next.delete(name)
  }
  next.delete('page')
  return next
}

export function useGuestFilters() {
  const [searchParams, setSearchParams] = useSearchParams()
  const filters = useMemo(() => parseGuestFilters(searchParams), [searchParams])

  const update = useCallback(
    (patch: Partial<Record<FilterKey, string>>) =>
      setSearchParams((current) => applyGuestFilters(current, patch)),
    [setSearchParams],
  )

  const setPage = useCallback(
    (page: number) =>
      setSearchParams((current) => {
        const next = new URLSearchParams(current)
        next.set('page', String(page))
        return next
      }),
    [setSearchParams],
  )

  const reset = useCallback(() => setSearchParams(new URLSearchParams()), [setSearchParams])

  const isFiltered = Boolean(
    filters.q || filters.sourceId || filters.tenantId || filters.periodFrom || filters.periodTo,
  )

  return { filters, update, setPage, reset, isFiltered }
}
