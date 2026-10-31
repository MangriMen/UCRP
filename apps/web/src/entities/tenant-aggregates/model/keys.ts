import type { Granularity } from './types'

export interface TenantAggregatesFilters {
  from: string
  to: string
  granularity?: Granularity
}

export const tenantAggregatesKeys = {
  all: ['tenant-aggregates'] as const,
  detail: (tenantId: string, filters: TenantAggregatesFilters) =>
    [...tenantAggregatesKeys.all, tenantId, filters] as const,
}
