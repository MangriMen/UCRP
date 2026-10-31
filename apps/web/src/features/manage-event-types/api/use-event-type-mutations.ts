import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventTypeKeys, type EventTypeRequest } from '@/entities/event-type'
import { api } from '@/shared/api/api'

function useInvalidateEventTypes() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: eventTypeKeys.all })
}

export function useCreateEventType() {
  const invalidate = useInvalidateEventTypes()
  return useMutation({
    mutationFn: (body: EventTypeRequest) => api.post('/event-types', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateEventType() {
  const invalidate = useInvalidateEventTypes()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: EventTypeRequest }) =>
      api.patch('/event-types/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useDeleteEventType() {
  const invalidate = useInvalidateEventTypes()
  return useMutation({
    mutationFn: (id: string) => api.delete('/event-types/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}
