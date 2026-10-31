import { useMutation, useQueryClient } from '@tanstack/react-query'
import { mallKeys, type MallCreateRequest, type MallUpdateRequest } from '@/entities/mall'
import { api } from '@/shared/api/api'

function useInvalidateMalls() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: mallKeys.all })
}

export function useCreateMall() {
  const invalidate = useInvalidateMalls()
  return useMutation({
    mutationFn: (body: MallCreateRequest) => api.post('/malls', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateMall() {
  const invalidate = useInvalidateMalls()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: MallUpdateRequest }) =>
      api.patch('/malls/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useDeleteMall() {
  const invalidate = useInvalidateMalls()
  return useMutation({
    mutationFn: (id: string) => api.delete('/malls/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}
