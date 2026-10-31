import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { dashboardKeys, type MetricsFilters, type PenetrationFilters } from '../model/keys'

export function useDashboardMetrics(filters: MetricsFilters, enabled = true) {
  return useQuery({
    queryKey: dashboardKeys.metrics(filters),
    queryFn: () => api.get('/dashboard/metrics', { query: filters }),
    enabled,
  })
}

export function usePenetration(filters: PenetrationFilters, enabled = true) {
  return useQuery({
    queryKey: dashboardKeys.penetration(filters),
    queryFn: () => api.get('/dashboard/penetration', { query: filters }),
    enabled,
  })
}
