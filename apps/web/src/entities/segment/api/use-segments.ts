import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { pollWhile, DEFAULT_POLL_MS } from '@/shared/api/polling'
import { segmentKeys, type SegmentListFilters, type SegmentMemberFilters } from '../model/keys'
import { isSegmentComputing } from '../model/status'
import type { Segment } from '../model/types'

export function useSegments(filters: SegmentListFilters = {}) {
  return useQuery({
    queryKey: segmentKeys.list(filters),
    queryFn: () =>
      api.get('/segments', {
        query: {
          status: filters.status,
          q: filters.query || undefined,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
    refetchInterval: pollWhile(
      (data) => data.items.some((segment) => isSegmentComputing(segment.status)),
      DEFAULT_POLL_MS,
    ),
  })
}

export function useSegment(id: string) {
  return useQuery<Segment>({
    queryKey: segmentKeys.detail(id),
    queryFn: () => api.get('/segments/{id}', { params: { id } }),
    refetchInterval: pollWhile((segment) => isSegmentComputing(segment.status), DEFAULT_POLL_MS),
  })
}

export function useSegmentMembers(id: string, filters: SegmentMemberFilters = {}) {
  return useQuery({
    queryKey: segmentKeys.members(id, filters),
    queryFn: () =>
      api.get('/segments/{id}/members', {
        params: { id },
        query: { version: filters.version, page: filters.page, page_size: filters.pageSize },
      }),
    placeholderData: keepPreviousData,
  })
}

export function useSegmentVersions(id: string) {
  return useQuery({
    queryKey: segmentKeys.versions(id),
    queryFn: () => api.get('/segments/{id}/versions', { params: { id } }),
  })
}

export function useSegmentTemplates() {
  return useQuery({
    queryKey: segmentKeys.templates(),
    queryFn: () => api.get('/segment-templates'),
  })
}
