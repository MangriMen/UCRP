import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  tenantCategoryKeys,
  type TenantCategoryRelation,
  type TenantCategoryRequest,
} from '@/entities/tenant-category'
import { api } from '@/shared/api/api'

function useInvalidateCategories() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: tenantCategoryKeys.all })
}

export function useCreateTenantCategory() {
  const invalidate = useInvalidateCategories()
  return useMutation({
    mutationFn: (body: TenantCategoryRequest) => api.post('/tenant-categories', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateTenantCategory() {
  const invalidate = useInvalidateCategories()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: TenantCategoryRequest }) =>
      api.patch('/tenant-categories/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useReplaceTenantCategoryRelations(categoryId: string) {
  const invalidate = useInvalidateCategories()
  return useMutation({
    mutationFn: (relations: TenantCategoryRelation[]) =>
      api.put('/tenant-categories/{id}/relations', {
        params: { id: categoryId },
        body: { relations },
      }),
    onSuccess: invalidate,
  })
}

export function useDeleteTenantCategory() {
  const invalidate = useInvalidateCategories()
  return useMutation({
    mutationFn: (id: string) => api.delete('/tenant-categories/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}
