import { useMutation } from '@tanstack/react-query'
import type { UserInvitationRequest } from '@/entities/user'
import { api } from '@/shared/api/api'

export function useInviteUser() {
  return useMutation({
    mutationFn: (body: UserInvitationRequest) => api.post('/users/invitations', { body }),
  })
}
