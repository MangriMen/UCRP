import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { mergeCandidateKeys, type MergeCandidateFilters } from '../model/keys'

export function useMergeCandidates(filters: MergeCandidateFilters) {
  return useQuery({
    queryKey: mergeCandidateKeys.list(filters),
    queryFn: () =>
      api.get('/merge-candidates', {
        query: {
          resolution: filters.resolution,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}

export function useMergeCandidate(id: string) {
  return useQuery({
    queryKey: mergeCandidateKeys.detail(id),
    queryFn: () => api.get('/merge-candidates/{id}', { params: { id } }),
  })
}
