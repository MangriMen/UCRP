import { useMutation, useQueryClient } from '@tanstack/react-query'
import { guestKeys } from '@/entities/guest'
import { mergeCandidateKeys } from '@/entities/merge-candidate'
import { unmergeBlockKeys } from '@/entities/unmerge-block'
import { api } from '@/shared/api/api'

export function useConfirmMerge() {
  const queryClient = useQueryClient()
  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: mergeCandidateKeys.all }),
      queryClient.invalidateQueries({ queryKey: unmergeBlockKeys.all }),
      queryClient.invalidateQueries({ queryKey: guestKeys.all }),
    ])
  return useMutation({
    mutationFn: (candidateId: string) =>
      api.post('/merge-candidates/{id}/confirm', { params: { id: candidateId } }),
    onSettled: refresh,
  })
}
