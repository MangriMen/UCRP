import { useMutation } from '@tanstack/react-query'
import type { UserInvitationRequest } from '@/entities/user'
import { requestBlob } from '@/shared/api/client'

export function useInviteUser() {
  return useMutation({
    mutationFn: async (body: UserInvitationRequest) => {
      await requestBlob('/users/invitations', { method: 'POST', body })
    },
  })
}
