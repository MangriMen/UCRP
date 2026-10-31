import type { AccessSubject, GrantStatus, GrantSubjectType } from './types'

export interface AccessGrantFilters {
  subjectType?: GrantSubjectType
  subjectId?: string
  status?: GrantStatus
  page?: number
  pageSize?: number
}

export interface AccessAuditFilters {
  grantId?: string
  actor?: string
  from?: string
  to?: string
  page?: number
  pageSize?: number
}

export const accessKeys = {
  all: ['access'] as const,
  grantList: (filters: AccessGrantFilters) => [...accessKeys.all, 'grants', filters] as const,
  grant: (id: string) => [...accessKeys.all, 'grant', id] as const,
  permissions: () => [...accessKeys.all, 'permissions'] as const,
  presets: () => [...accessKeys.all, 'presets'] as const,
  effective: (subject: AccessSubject | null) => [...accessKeys.all, 'effective', subject] as const,
  audit: (filters: AccessAuditFilters) => [...accessKeys.all, 'audit', filters] as const,
}
