import { useCallback, useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { parsePage } from './search-params'

export const DEFAULT_PAGE_SIZE = 50
export const PAGE_SIZE_OPTIONS = [20, 50, 100] as const

export type SortDirection = 'asc' | 'desc'

export interface SortState {
  field: string
  direction: SortDirection
}

export type ParamChanges = Record<string, string | null | undefined>

export const parsePageSize = (value: string | null) => {
  const size = Number(value)
  return PAGE_SIZE_OPTIONS.some((option) => option === size) ? size : DEFAULT_PAGE_SIZE
}

export const parseSort = (value: string | null): SortState | undefined => {
  if (!value) return undefined
  const descending = value.startsWith('-')
  const field = descending ? value.slice(1) : value
  return field ? { field, direction: descending ? 'desc' : 'asc' } : undefined
}

export const formatSort = (sort: SortState | undefined) =>
  sort ? `${sort.direction === 'desc' ? '-' : ''}${sort.field}` : undefined

export const nextSort = (current: SortState | undefined, field: string): SortState | undefined => {
  if (current?.field !== field) return { field, direction: 'asc' }
  return current.direction === 'asc' ? { field, direction: 'desc' } : undefined
}

export function applyParamChanges(
  current: URLSearchParams,
  changes: ParamChanges,
  resetPage = true,
): URLSearchParams {
  const next = new URLSearchParams(current)
  for (const [key, value] of Object.entries(changes)) {
    if (value === null || value === undefined || value === '') next.delete(key)
    else next.set(key, value)
  }
  if (resetPage && !('page' in changes)) next.delete('page')
  return next
}

export type ListParams = ReturnType<typeof useListParams>

export function useListParams() {
  const [searchParams, setSearchParams] = useSearchParams()

  const page = parsePage(searchParams.get('page'))
  const pageSize = parsePageSize(searchParams.get('page_size'))
  const sortValue = searchParams.get('sort')
  const sort = useMemo(() => parseSort(sortValue), [sortValue])

  const get = useCallback((key: string) => searchParams.get(key) ?? '', [searchParams])

  const update = useCallback(
    (changes: ParamChanges) =>
      setSearchParams((current) => applyParamChanges(current, changes), { replace: true }),
    [setSearchParams],
  )

  const setPage = useCallback(
    (next: number) =>
      setSearchParams((current) => applyParamChanges(current, { page: String(next) }, false)),
    [setSearchParams],
  )

  const setPageSize = useCallback(
    (size: number) => update({ page_size: size === DEFAULT_PAGE_SIZE ? null : String(size) }),
    [update],
  )

  const setSort = useCallback(
    (next: SortState | undefined) => update({ sort: formatSort(next) }),
    [update],
  )

  const isFiltered = useCallback(
    (keys: readonly string[]) => keys.some((key) => get(key) !== ''),
    [get],
  )

  const reset = useCallback(() => setSearchParams(new URLSearchParams()), [setSearchParams])

  return {
    page,
    pageSize,
    sort,
    get,
    update,
    setPage,
    setPageSize,
    setSort,
    isFiltered,
    reset,
  }
}
