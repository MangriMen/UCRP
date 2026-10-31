import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { tenantCategoryKeys } from '../model/keys'

export function useTenantCategories() {
  return useQuery({
    queryKey: tenantCategoryKeys.all,
    queryFn: () => api.get('/tenant-categories'),
  })
}
