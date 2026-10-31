import { useMutation, useQueryClient } from '@tanstack/react-query'
import { tenantKeys, type TenantCreateRequest, type TenantUpdateRequest } from '@/entities/tenant'
import { api } from '@/shared/api/api'

function useInvalidateTenants() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: tenantKeys.all })
}

export function useCreateTenant() {
  const invalidate = useInvalidateTenants()
  return useMutation({
    mutationFn: (body: TenantCreateRequest) => api.post('/tenants', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateTenant() {
  const invalidate = useInvalidateTenants()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: TenantUpdateRequest }) =>
      api.patch('/tenants/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useDeleteTenant() {
  const invalidate = useInvalidateTenants()
  return useMutation({
    mutationFn: (id: string) => api.delete('/tenants/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}
