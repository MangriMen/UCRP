import type { MergeResolution } from './types'

export interface MergeCandidateFilters {
  resolution?: MergeResolution
  page?: number
  pageSize?: number
}

export const mergeCandidateKeys = {
  all: ['merge-candidates'] as const,
  lists: () => [...mergeCandidateKeys.all, 'list'] as const,
  list: (filters: MergeCandidateFilters) => [...mergeCandidateKeys.lists(), filters] as const,
  detail: (id: string) => [...mergeCandidateKeys.all, 'detail', id] as const,
}
