import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { endSession } from '@/entities/user'
import { api } from '@/shared/api/api'

export function useLogout() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  return useMutation({
    mutationFn: () => api.post('/auth/logout'),
    onSettled: () => {
      endSession(queryClient)
      navigate('/login', { replace: true, state: null })
    },
  })
}
