import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { tenantAggregatesKeys } from '../model/keys'

export function useTenantAggregates(tenantId: string | null) {
  return useQuery({
    queryKey: tenantAggregatesKeys.detail(tenantId ?? ''),
    queryFn: () => api.get('/tenants/{id}/aggregates', { params: { id: tenantId ?? '' } }),
    enabled: tenantId !== null,
  })
}
