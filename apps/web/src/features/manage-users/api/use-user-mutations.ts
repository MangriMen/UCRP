import { useMutation, useQueryClient } from '@tanstack/react-query'
import { userKeys, type UserCreateRequest, type UserUpdateRequest } from '@/entities/user'
import { api } from '@/shared/api/api'

function useInvalidateUsers() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: userKeys.all })
}

export function useCreateUser() {
  const invalidate = useInvalidateUsers()
  return useMutation({
    mutationFn: (body: UserCreateRequest) => api.post('/users', { body }),
    onSuccess: invalidate,
  })
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
