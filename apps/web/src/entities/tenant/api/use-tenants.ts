import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { hasKey, narrow } from '@/shared/api/narrow'
import { tenantKeys, type TenantListFilters } from '../model/keys'

export function useTenants(filters: TenantListFilters = {}) {
  return useQuery({
    queryKey: tenantKeys.list(filters),
    queryFn: () =>
      api.get('/tenants', {
        query: { q: filters.q, page: filters.page, page_size: filters.pageSize },
      }),
    select: (data) => narrow(data, hasKey('meta')),
    placeholderData: keepPreviousData,
  })
}
