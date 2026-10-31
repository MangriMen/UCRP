import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { tenantAggregatesKeys, type TenantAggregatePeriod } from '../model/keys'

export function useTenantAggregates(tenantId: string | null, period?: TenantAggregatePeriod) {
  return useQuery({
    queryKey: tenantAggregatesKeys.detail(tenantId ?? '', period),
    queryFn: () =>
      api.get('/tenants/{id}/aggregates', {
        params: { id: tenantId ?? '' },
        query: period,
      }),
    enabled: tenantId !== null,
    placeholderData: keepPreviousData,
  })
}
