import { useMutation, useQueryClient } from '@tanstack/react-query'
import { userKeys } from '@/entities/user'
import { api } from '@/shared/api/api'
import { setAccessToken } from '@/shared/api/client'

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (credentials: { email: string; password: string }) =>
      api.post('/auth/login', { body: credentials }),
    onSuccess: ({ accessToken, user }) => {
      setAccessToken(accessToken)
      queryClient.setQueryData(userKeys.me, user)
    },
  })
}
