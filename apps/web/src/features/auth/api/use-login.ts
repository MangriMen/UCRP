import { useMutation, useQueryClient } from '@tanstack/react-query'
import { startSession } from '@/entities/user'
import { api } from '@/shared/api/api'
import { ApiError } from '@/shared/api/client'

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (credentials: { email: string; password: string }) => {
      const session = await api.post('/auth/login', { body: credentials })
      if (!session.accessToken || !session.user) {
        throw new ApiError(502, {
          code: 'internal_error',
          message: 'Сервер вернул некорректный ответ при входе',
        })
      }
      return session
    },
    onSuccess: ({ accessToken, user }) => startSession(queryClient, accessToken, user),
  })
}
