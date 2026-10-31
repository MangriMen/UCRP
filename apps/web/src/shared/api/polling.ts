import type { Query } from '@tanstack/react-query'

export const IMPORT_POLL_MS = 2000

export function pollWhile<TData>(isActive: (data: TData) => boolean, intervalMs: number) {
  return (query: Query<TData, Error, TData>) => {
    const data = query.state.data
    return data !== undefined && isActive(data) ? intervalMs : false
  }
}

export const SEGMENT_POLL_MS = 5000
