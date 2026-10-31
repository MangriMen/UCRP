export interface TenantListFilters {
  q?: string
  categoryId?: string
  floor?: number
  isActive?: boolean
  periodFrom?: string
  periodTo?: string
  sort?: string
  page?: number
  pageSize?: number
}

export const tenantKeys = {
  all: ['tenants'] as const,
  lists: () => [...tenantKeys.all, 'list'] as const,
  list: (filters: TenantListFilters) => [...tenantKeys.lists(), filters] as const,
  detail: (id: string) => [...tenantKeys.all, 'detail', id] as const,
}
