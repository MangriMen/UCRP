import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { mallKeys } from '../model/keys'

export function useMalls() {
  return useQuery({
    queryKey: mallKeys.all,
    queryFn: () => api.get('/malls'),
  })
}
