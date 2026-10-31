import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { contactLimitKeys } from '../model/keys'

export function useContactLimits() {
  return useQuery({
    queryKey: contactLimitKeys.all,
    queryFn: () => api.get('/contact-limits'),
  })
}
