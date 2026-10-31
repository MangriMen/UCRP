import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { mergeBlockKeys, type MergeBlockFilters } from '../model/keys'

export function useMergeBlocks(filters: MergeBlockFilters = {}) {
  return useQuery({
    queryKey: mergeBlockKeys.list(filters),
    queryFn: () =>
      api.get('/merge-blocks', { query: { page: filters.page, page_size: filters.pageSize } }),
    placeholderData: keepPreviousData,
  })
}
