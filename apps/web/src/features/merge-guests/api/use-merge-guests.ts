import { useMutation, useQueryClient } from '@tanstack/react-query'
import { guestKeys } from '@/entities/guest'
import { mergeCandidateKeys } from '@/entities/merge-candidate'
import { api } from '@/shared/api/api'

interface MergeGuestsInput {
  guestId: string
  targetGuestId: string
}

export function useMergeGuests() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ guestId, targetGuestId }: MergeGuestsInput) =>
      api.post('/guests/{id}/merge', { params: { id: guestId }, body: { targetGuestId } }),
    onSettled: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: guestKeys.all }),
        queryClient.invalidateQueries({ queryKey: mergeCandidateKeys.all }),
      ]),
  })
}
