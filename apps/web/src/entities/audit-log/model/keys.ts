export interface AuditLogFilters {
  entityType?: string
  entityId?: string
  actor?: string
  action?: string
  from?: string
  to?: string
  page?: number
  pageSize?: number
}

export const auditLogKeys = {
  all: ['audit-logs'] as const,
  list: (filters: AuditLogFilters) => [...auditLogKeys.all, 'list', filters] as const,
}
