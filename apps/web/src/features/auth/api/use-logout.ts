import { useMutation, useQueryClient } from '@tanstack/react-query'
import { endSession } from '@/entities/user'
import { api } from '@/shared/api/api'

export function useLogout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post('/auth/logout'),
    onSettled: () => endSession(queryClient),
  })
}
