import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { importKeys } from '../model/keys'

export function useRejectedRows(id: string, page = 1, pageSize = 50) {
  return useQuery({
    queryKey: importKeys.rejectedRows(id, page, pageSize),
    queryFn: () =>
      api.get('/imports/{id}/rejected-rows', {
        params: { id },
        query: { page, page_size: pageSize },
      }),
    placeholderData: keepPreviousData,
  })
}
