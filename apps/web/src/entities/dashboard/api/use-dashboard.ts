import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { dashboardKeys, type MetricsFilters, type CrossTenantFilters } from '../model/keys'

export function useDashboardMetrics(filters: MetricsFilters, enabled = true) {
  return useQuery({
    queryKey: dashboardKeys.metrics(filters),
    queryFn: () => api.get('/dashboard/metrics', { query: filters }),
    enabled,
  })
}

export function useCrossTenantReport(filters: CrossTenantFilters, enabled = true) {
  return useQuery({
    queryKey: dashboardKeys.crossTenant(filters),
    queryFn: () => api.get('/dashboard/cross-tenant', { query: filters }),
    enabled,
  })
}
