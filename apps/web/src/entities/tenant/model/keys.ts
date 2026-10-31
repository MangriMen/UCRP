export interface TenantListFilters {
  q?: string
  page?: number
  pageSize?: number
}

export const tenantKeys = {
  all: ['tenants'] as const,
  lists: () => [...tenantKeys.all, 'list'] as const,
  list: (filters: TenantListFilters) => [...tenantKeys.lists(), filters] as const,
}
