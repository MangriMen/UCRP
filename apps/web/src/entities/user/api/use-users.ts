import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { userKeys, type UserListFilters } from '../model/keys'

export function useUsers(filters: UserListFilters = {}) {
  return useQuery({
    queryKey: userKeys.list(filters),
    queryFn: () =>
      api.get('/users', {
        query: {
          q: filters.q,
          role: filters.role,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
  })
}
