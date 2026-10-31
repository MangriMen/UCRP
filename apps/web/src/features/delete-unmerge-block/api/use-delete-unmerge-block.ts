import { useMutation, useQueryClient } from '@tanstack/react-query'
import { unmergeBlockKeys } from '@/entities/unmerge-block'
import { api } from '@/shared/api/api'

export function useDeleteUnmergeBlock() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (blockId: string) =>
      api.delete('/unmerge-blocks/{id}', { params: { id: blockId } }),
    onSettled: () => queryClient.invalidateQueries({ queryKey: unmergeBlockKeys.all }),
  })
}
