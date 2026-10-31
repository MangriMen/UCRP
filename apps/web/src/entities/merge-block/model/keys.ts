export interface MergeBlockFilters {
  page?: number
  pageSize?: number
}

export const mergeBlockKeys = {
  all: ['merge-blocks'] as const,
  list: (filters: MergeBlockFilters) => [...mergeBlockKeys.all, 'list', filters] as const,
}
