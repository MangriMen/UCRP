import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { tenantCategoryKeys } from '../model/keys'

export function useTenantCategoryRelations(categoryId: string) {
  return useQuery({
    queryKey: tenantCategoryKeys.relations(categoryId),
    queryFn: () => api.get('/tenant-categories/{id}/relations', { params: { id: categoryId } }),
  })
}
