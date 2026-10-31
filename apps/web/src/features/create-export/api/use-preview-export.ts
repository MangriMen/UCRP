import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { ExportChannel } from '@/entities/export'
import { api } from '@/shared/api/api'

interface PreviewParams {
  segmentId: string
  channel: ExportChannel
  controlPercent: number
}

export function usePreviewExport(params: PreviewParams, enabled: boolean) {
  return useQuery({
    queryKey: ['exports', 'preview', params],
    queryFn: () => api.post('/exports/preview', { body: params }),
    enabled,
    placeholderData: keepPreviousData,
  })
}
