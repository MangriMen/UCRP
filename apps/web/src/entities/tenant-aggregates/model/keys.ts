export interface TenantAggregatePeriod {
  from: string
  to: string
}

export const tenantAggregatesKeys = {
  all: ['tenant-aggregates'] as const,
  detail: (tenantId: string, period?: TenantAggregatePeriod) =>
    [...tenantAggregatesKeys.all, tenantId, period ?? null] as const,
}
