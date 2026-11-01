import type { Query } from '@tanstack/react-query'

export const IMPORT_POLL_MS = 2000

export function pollWhile<TData>(
  isActive: (data: TData) => boolean,
  intervalMs: number,
  maxUpdates = Infinity,
) {
  return (query: Query<TData, Error, TData>) => {
    const data = query.state.data
    if (data === undefined || query.state.dataUpdateCount >= maxUpdates) return false
    return isActive(data) ? intervalMs : false
  }
}

export const DEFAULT_POLL_MS = 5000
