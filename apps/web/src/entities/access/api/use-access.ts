import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { accessKeys, type AccessAuditFilters, type AccessGrantFilters } from '../model/keys'
import type { AccessSubject } from '../model/types'

const CATALOG_STALE_MS = 5 * 60_000

export function useAccessGrants(filters: AccessGrantFilters = {}) {
  return useQuery({
    queryKey: accessKeys.grantList(filters),
    queryFn: () =>
      api.get('/access/grants', {
        query: {
          subject_type: filters.subjectType,
          subject_id: filters.subjectId,
          status: filters.status,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}

export function useAccessGrant(id: string) {
  return useQuery({
    queryKey: accessKeys.grant(id),
    queryFn: () => api.get('/access/grants/{id}', { params: { id } }),
  })
}

export function useAccessPermissions() {
  return useQuery({
    queryKey: accessKeys.permissions(),
    queryFn: () => api.get('/access/permissions'),
    select: (data) => data.items,
    staleTime: CATALOG_STALE_MS,
  })
}

export function useAccessPresets() {
  return useQuery({
    queryKey: accessKeys.presets(),
    queryFn: () => api.get('/access/presets'),
    select: (data) => data.items,
    staleTime: CATALOG_STALE_MS,
  })
}

export function useEffectiveAccess(subject: AccessSubject | null) {
  return useQuery({
    queryKey: accessKeys.effective(subject),
    queryFn: () =>
      api.get('/access/effective', {
        query: {
          subject_type: subject?.subjectType ?? '',
          subject_id: subject?.subjectId ?? '',
        },
      }),
    enabled: subject !== null,
  })
}

export function useAccessAudit(filters: AccessAuditFilters) {
  return useQuery({
    queryKey: accessKeys.audit(filters),
    queryFn: () =>
      api.get('/access/audit', {
        query: {
          grant_id: filters.grantId,
          actor: filters.actor,
          from: filters.from,
          to: filters.to,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}
