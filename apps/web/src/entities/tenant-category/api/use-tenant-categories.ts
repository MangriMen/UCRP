import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { tenantCategoryKeys } from '../model/keys'

interface UseTenantCategoriesOptions {
  includeInactive?: boolean
}

export function useTenantCategories({ includeInactive = false }: UseTenantCategoriesOptions = {}) {
  return useQuery({
    queryKey: tenantCategoryKeys.list(includeInactive),
    queryFn: () =>
      api.get('/tenant-categories', {
        query: includeInactive ? { include_inactive: true } : undefined,
      }),
    placeholderData: keepPreviousData,
  })
}
