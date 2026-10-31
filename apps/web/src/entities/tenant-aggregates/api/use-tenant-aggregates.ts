import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { tenantAggregatesKeys, type TenantAggregatesFilters } from '../model/keys'

export function useTenantAggregates(tenantId: string | null, filters: TenantAggregatesFilters) {
  return useQuery({
    queryKey: tenantAggregatesKeys.detail(tenantId ?? '', filters),
    queryFn: () =>
      api.get('/tenants/{id}/aggregates', { params: { id: tenantId ?? '' }, query: filters }),
    enabled: tenantId !== null,
  })
}
