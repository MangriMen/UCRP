import { useMutation, useQueryClient } from '@tanstack/react-query'
import { userKeys, type UserUpdateRequest } from '@/entities/user'
import { api } from '@/shared/api/api'

function useInvalidateUsers() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: userKeys.all })
}

export function useUpdateUser() {
  const invalidate = useInvalidateUsers()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UserUpdateRequest }) =>
      api.patch('/users/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useDeleteUser() {
  const invalidate = useInvalidateUsers()
  return useMutation({
    mutationFn: (id: string) => api.delete('/users/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}

export function useRecoverUser() {
  return useMutation({
    mutationFn: (id: string) => api.post('/users/{id}/recovery', { params: { id } }),
  })
}
