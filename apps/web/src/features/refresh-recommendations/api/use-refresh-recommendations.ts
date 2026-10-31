import { useMutation, useQueryClient } from '@tanstack/react-query'
import { recommendationKeys } from '@/entities/recommendation'
import { api } from '@/shared/api/api'

export function useRefreshRecommendations() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (guestId: string) =>
      api.post('/guests/{id}/recommendations/refresh', { params: { id: guestId } }),
    onSuccess: (list, guestId) => queryClient.setQueryData(recommendationKeys.guest(guestId), list),
  })
}
