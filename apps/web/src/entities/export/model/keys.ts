import type { ExportChannel, ExportStatus } from './types'

export interface ExportListFilters {
  segmentId?: string
  campaignId?: string
  channel?: ExportChannel
  status?: ExportStatus
  page?: number
  pageSize?: number
}

export interface ExportMemberFilters {
  isControl: boolean
  page?: number
  pageSize?: number
}

export const exportKeys = {
  all: ['exports'] as const,
  lists: () => [...exportKeys.all, 'list'] as const,
  list: (filters: ExportListFilters) => [...exportKeys.lists(), filters] as const,
  detail: (id: string) => [...exportKeys.all, 'detail', id] as const,
  members: (id: string, filters: ExportMemberFilters) =>
    [...exportKeys.detail(id), 'members', filters] as const,
}
