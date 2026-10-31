import { useMutation, useQueryClient } from '@tanstack/react-query'
import { guestKeys } from '@/entities/guest'
import { api } from '@/shared/api/api'

export function useDeleteGuest() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (guestId: string) => api.delete('/guests/{id}', { params: { id: guestId } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: guestKeys.all }),
  })
}
