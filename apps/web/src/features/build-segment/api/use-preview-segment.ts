import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { segmentKeys, type SegmentRuleNode } from '@/entities/segment'
import { api } from '@/shared/api/api'

const PREVIEW_TIMEOUT_MS = 10_000
const PREVIEW_EXAMPLES_LIMIT = 5

export function usePreviewSegment(rule: SegmentRuleNode | null) {
  return useQuery({
    queryKey: [...segmentKeys.all, 'preview', rule],
    queryFn: ({ signal }) =>
      api.post('/segments/preview', {
        body: { rule: rule as SegmentRuleNode, limit: PREVIEW_EXAMPLES_LIMIT },
        signal: AbortSignal.any([signal, AbortSignal.timeout(PREVIEW_TIMEOUT_MS)]),
      }),
    enabled: rule !== null,
    placeholderData: keepPreviousData,
    retry: false,
    staleTime: Infinity,
  })
}
