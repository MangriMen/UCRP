import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { dataSourceKeys } from '../model/keys'

const MAX_PAGE_SIZE = 200

export function useDataSources() {
  return useQuery({
    queryKey: dataSourceKeys.list(),
    queryFn: () => api.get('/data-sources', { query: { page_size: MAX_PAGE_SIZE } }),
    select: (page) => page.items,
  })
}
