import type { Granularity } from './types'

export interface MetricsFilters {
  from: string
  to: string
}

export interface PenetrationFilters extends MetricsFilters {
  granularity?: Granularity
}

export const dashboardKeys = {
  all: ['dashboard'] as const,
  metrics: (filters: MetricsFilters) => [...dashboardKeys.all, 'metrics', filters] as const,
  penetration: (filters: PenetrationFilters) =>
    [...dashboardKeys.all, 'penetration', filters] as const,
}
