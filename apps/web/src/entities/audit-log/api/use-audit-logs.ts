import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { auditLogKeys, type AuditLogFilters } from '../model/keys'

export function useAuditLogs(filters: AuditLogFilters = {}) {
  return useQuery({
    queryKey: auditLogKeys.list(filters),
    queryFn: () =>
      api.get('/audit-logs', {
        query: {
          entity_type: filters.entityType,
          entity_id: filters.entityId,
          actor: filters.actor,
          action: filters.action,
          from: filters.from,
          to: filters.to,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}
