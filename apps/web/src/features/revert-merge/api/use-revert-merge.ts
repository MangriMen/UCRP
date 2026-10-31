import { useMutation, useQueryClient } from '@tanstack/react-query'
import { guestKeys } from '@/entities/guest'
import { mergeCandidateKeys } from '@/entities/merge-candidate'
import { mergeBlockKeys } from '@/entities/merge-block'
import { api } from '@/shared/api/api'

export function useRevertMerge() {
  const queryClient = useQueryClient()
  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: mergeCandidateKeys.all }),
      queryClient.invalidateQueries({ queryKey: mergeBlockKeys.all }),
      queryClient.invalidateQueries({ queryKey: guestKeys.all }),
    ])
  return useMutation({
    mutationFn: (candidateId: string) =>
      api.post('/merge-candidates/{id}/revert', { params: { id: candidateId } }),
    onSettled: refresh,
  })
}
