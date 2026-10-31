import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  segmentKeys,
  type SegmentCreateRequest,
  type SegmentUpdateRequest,
} from '@/entities/segment'
import { api } from '@/shared/api/api'

export function useCreateSegment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: SegmentCreateRequest) => api.post('/segments', { body }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: segmentKeys.all }),
  })
}

export function useUpdateSegment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: SegmentUpdateRequest }) =>
      api.put('/segments/{id}', { params: { id }, body }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: segmentKeys.all }),
  })
}
