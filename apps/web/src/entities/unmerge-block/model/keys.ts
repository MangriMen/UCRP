export interface UnmergeBlockFilters {
  page?: number
  pageSize?: number
}

export const unmergeBlockKeys = {
  all: ['unmerge-blocks'] as const,
  list: (filters: UnmergeBlockFilters) => [...unmergeBlockKeys.all, 'list', filters] as const,
}
