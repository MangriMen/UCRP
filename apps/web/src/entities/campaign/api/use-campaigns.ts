import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { campaignKeys, type CampaignListFilters } from '../model/keys'
import type { Campaign, CampaignResults } from '../model/types'

export function useCampaigns(filters: CampaignListFilters = {}) {
  return useQuery({
    queryKey: campaignKeys.list(filters),
    queryFn: () =>
      api.get('/campaigns', {
        query: {
          status: filters.status,
          segment_id: filters.segmentId,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}

export function useCampaign(id: string) {
  return useQuery<Campaign>({
    queryKey: campaignKeys.detail(id),
    queryFn: () => api.get('/campaigns/{id}', { params: { id } }),
  })
}

export function useCampaignResults(id: string, enabled = true) {
  return useQuery<CampaignResults>({
    queryKey: campaignKeys.results(id),
    queryFn: () => api.get('/campaigns/{id}/results', { params: { id } }),
    enabled,
  })
}
