import type { SegmentStatus } from './types'

export interface SegmentListFilters {
  status?: SegmentStatus
  query?: string
  page?: number
  pageSize?: number
}

export interface SegmentMemberFilters {
  version?: number
  page?: number
  pageSize?: number
}

export const segmentKeys = {
  all: ['segments'] as const,
  lists: () => [...segmentKeys.all, 'list'] as const,
  list: (filters: SegmentListFilters) => [...segmentKeys.lists(), filters] as const,
  detail: (id: string) => [...segmentKeys.all, 'detail', id] as const,
  members: (id: string, filters: SegmentMemberFilters) =>
    [...segmentKeys.detail(id), 'members', filters] as const,
  versions: (id: string) => [...segmentKeys.detail(id), 'versions'] as const,
  templates: () => [...segmentKeys.all, 'templates'] as const,
}
