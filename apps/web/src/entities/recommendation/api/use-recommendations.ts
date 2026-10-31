import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { pollWhile, DEFAULT_POLL_MS } from '@/shared/api/polling'
import { recommendationKeys, SEGMENT_TOP_TENANTS_LIMIT } from '../model/keys'
import type { RecommendationList, SegmentRecommendationList } from '../model/types'

export function useGuestRecommendations(guestId: string) {
  return useQuery<RecommendationList>({
    queryKey: recommendationKeys.guest(guestId),
    queryFn: () => api.get('/guests/{id}/recommendations', { params: { id: guestId } }),
    refetchInterval: pollWhile((list) => !list.computedAt, DEFAULT_POLL_MS),
  })
}

export function useSegmentRecommendations(
  segmentId: string,
  isSegmentComputing: boolean,
  limit = SEGMENT_TOP_TENANTS_LIMIT,
) {
  return useQuery<SegmentRecommendationList>({
    queryKey: recommendationKeys.segment(segmentId, limit),
    queryFn: () =>
      api.get('/segments/{id}/recommendations', { params: { id: segmentId }, query: { limit } }),
    refetchInterval: pollWhile(() => isSegmentComputing, DEFAULT_POLL_MS),
  })
}
