export const tenantAggregatesKeys = {
  all: ['tenant-aggregates'] as const,
  detail: (tenantId: string) => [...tenantAggregatesKeys.all, tenantId] as const,
}
