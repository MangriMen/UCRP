import type { ExportChannel, ExportStatus } from './types'

export interface ExportListFilters {
  segmentId?: string
  channel?: ExportChannel
  status?: ExportStatus
  page?: number
  pageSize?: number
}

export const exportKeys = {
  all: ['exports'] as const,
  lists: () => [...exportKeys.all, 'list'] as const,
  list: (filters: ExportListFilters) => [...exportKeys.lists(), filters] as const,
  detail: (id: string) => [...exportKeys.all, 'detail', id] as const,
}
