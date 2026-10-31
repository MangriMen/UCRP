import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { unmergeBlockKeys, type UnmergeBlockFilters } from '../model/keys'

export function useUnmergeBlocks(filters: UnmergeBlockFilters = {}) {
  return useQuery({
    queryKey: unmergeBlockKeys.list(filters),
    queryFn: () =>
      api.get('/unmerge-blocks', { query: { page: filters.page, page_size: filters.pageSize } }),
    placeholderData: keepPreviousData,
  })
}
