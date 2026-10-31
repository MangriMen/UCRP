import { useMutation, useQueryClient } from '@tanstack/react-query'
import { startSession } from '@/entities/user'
import { api } from '@/shared/api/api'

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (credentials: { email: string; password: string }) =>
      api.post('/auth/login', { body: credentials }),
    onSuccess: ({ accessToken, user }) => startSession(queryClient, accessToken, user),
  })
}
