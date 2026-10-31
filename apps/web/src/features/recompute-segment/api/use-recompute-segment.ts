import { useMutation, useQueryClient } from '@tanstack/react-query'
import { segmentKeys } from '@/entities/segment'
import { api } from '@/shared/api/api'

export function useRecomputeSegment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (segmentId: string) =>
      api.post('/segments/{id}/recompute', { params: { id: segmentId } }),
    onSettled: () => queryClient.invalidateQueries({ queryKey: segmentKeys.all }),
  })
}
