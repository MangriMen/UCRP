import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { pollWhile, DEFAULT_POLL_MS } from '@/shared/api/polling'
import { exportKeys, type ExportListFilters, type ExportMemberFilters } from '../model/keys'
import { isExportActive } from '../model/status'

export function useExports(filters: ExportListFilters = {}) {
  return useQuery({
    queryKey: exportKeys.list(filters),
    queryFn: () =>
      api.get('/exports', {
        query: {
          segment_id: filters.segmentId,
          campaign_id: filters.campaignId,
          channel: filters.channel,
          status: filters.status,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
    refetchInterval: pollWhile(
      (data) => data.items.some((item) => isExportActive(item.status)),
      DEFAULT_POLL_MS,
    ),
  })
}

export function useExport(id: string) {
  return useQuery({
    queryKey: exportKeys.detail(id),
    queryFn: () => api.get('/exports/{id}', { params: { id } }),
    refetchInterval: pollWhile((data) => isExportActive(data.status), DEFAULT_POLL_MS),
  })
}

export function useExportMembers(id: string, filters: ExportMemberFilters) {
  return useQuery({
    queryKey: exportKeys.members(id, filters),
    queryFn: () =>
      api.get('/exports/{id}/members', {
        params: { id },
        query: {
          is_control: filters.isControl,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}
