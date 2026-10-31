import { useMutation } from '@tanstack/react-query'
import type { InvitationAcceptRequest } from '@/entities/user'
import { api } from '@/shared/api/api'

export function useAcceptInvitation() {
  return useMutation({
    mutationFn: (body: InvitationAcceptRequest) => api.post('/auth/invitations/accept', { body }),
  })
}
