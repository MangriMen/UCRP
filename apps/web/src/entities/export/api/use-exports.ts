import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { pollWhile, SEGMENT_POLL_MS } from '@/shared/api/polling'
import { exportKeys, type ExportListFilters } from '../model/keys'
import { isExportActive } from '../model/status'

export function useExports(filters: ExportListFilters = {}) {
  return useQuery({
    queryKey: exportKeys.list(filters),
    queryFn: () =>
      api.get('/exports', {
        query: {
          segment_id: filters.segmentId,
          channel: filters.channel,
          status: filters.status,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
    refetchInterval: pollWhile(
      (data) => data.items.some((item) => isExportActive(item.status)),
      SEGMENT_POLL_MS,
    ),
  })
}
