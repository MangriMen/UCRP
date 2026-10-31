import { useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { dataSourceKeys, type DataSourceListFilters } from '../model/keys'

const MAX_PAGE_SIZE = 200

export function useDataSources(filters: DataSourceListFilters = {}) {
  return useQuery({
    queryKey: dataSourceKeys.list(filters),
    queryFn: () =>
      api.get('/data-sources', {
        query: {
          tenant_id: filters.tenantId,
          type_id: filters.typeId,
          is_active: filters.isActive,
          page_size: MAX_PAGE_SIZE,
        },
      }),
    select: (page) => page.items,
  })
}
