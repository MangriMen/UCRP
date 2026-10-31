import type { CampaignStatus } from './types'

export interface CampaignListFilters {
  status?: CampaignStatus
  segmentId?: string
  page?: number
  pageSize?: number
}

export const campaignKeys = {
  all: ['campaigns'] as const,
  lists: () => [...campaignKeys.all, 'list'] as const,
  list: (filters: CampaignListFilters) => [...campaignKeys.lists(), filters] as const,
  detail: (id: string) => [...campaignKeys.all, 'detail', id] as const,
  results: (id: string) => [...campaignKeys.detail(id), 'results'] as const,
}
