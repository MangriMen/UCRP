import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { eventTypeKeys } from '../model/keys'

export function useEventTypes() {
  return useQuery({
    queryKey: eventTypeKeys.all,
    queryFn: () => api.get('/event-types'),
  })
}
