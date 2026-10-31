import { useMutation, useQueryClient } from '@tanstack/react-query'
import { mergeBlockKeys } from '@/entities/merge-block'
import { api } from '@/shared/api/api'

export function useDeleteMergeBlock() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (blockId: string) => api.delete('/merge-blocks/{id}', { params: { id: blockId } }),
    onSettled: () => queryClient.invalidateQueries({ queryKey: mergeBlockKeys.all }),
  })
}
