import type { Granularity } from './types'

export interface MetricsFilters {
  from: string
  to: string
}

export interface CrossTenantFilters extends MetricsFilters {
  granularity?: Granularity
}

export const dashboardKeys = {
  all: ['dashboard'] as const,
  metrics: (filters: MetricsFilters) => [...dashboardKeys.all, 'metrics', filters] as const,
  crossTenant: (filters: CrossTenantFilters) =>
    [...dashboardKeys.all, 'crossTenant', filters] as const,
}
