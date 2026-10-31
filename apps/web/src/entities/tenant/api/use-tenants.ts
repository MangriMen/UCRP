import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { hasKey, narrow } from '@/shared/api/narrow'
import { tenantKeys, type TenantListFilters } from '../model/keys'
import type { Tenant, TenantListSort } from '../model/types'

export function useTenants(filters: TenantListFilters = {}) {
  return useQuery({
    queryKey: tenantKeys.list(filters),
    queryFn: () =>
      api.get('/tenants', {
        query: {
          q: filters.q,
          category_id: filters.categoryId,
          floor: filters.floor,
          is_active: filters.isActive,
          period_from: filters.periodFrom,
          period_to: filters.periodTo,
          sort: filters.sort as TenantListSort | undefined,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    select: (data) => narrow(data, hasKey('meta')),
    placeholderData: keepPreviousData,
  })
}

export function useTenant(id: string) {
  return useQuery({
    queryKey: tenantKeys.detail(id),
    queryFn: () => api.get('/tenants/{id}', { params: { id } }),
    select: (data): Tenant => narrow(data, hasKey('mallId')),
  })
}
