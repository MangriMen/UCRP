import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { dataSourceKeys, type DataSourceListFilters } from '../model/keys'
import type { DataSourceImportStatus, DataSourceSort } from '../model/types'

const OPTIONS_PAGE_SIZE = 200

const fetchDataSources = (filters: DataSourceListFilters) =>
  api.get('/data-sources', {
    query: {
      tenant_id: filters.tenantId,
      type_id: filters.typeId,
      is_active: filters.isActive,
      q: filters.q,
      import_status: filters.importStatus as DataSourceImportStatus | undefined,
      sort: filters.sort as DataSourceSort | undefined,
      page: filters.page,
      page_size: filters.pageSize,
    },
  })

export function useDataSources(filters: DataSourceListFilters = {}) {
  return useQuery({
    queryKey: dataSourceKeys.list(filters),
    queryFn: () => fetchDataSources(filters),
    placeholderData: keepPreviousData,
  })
}

export function useDataSourceOptions() {
  const filters: DataSourceListFilters = { pageSize: OPTIONS_PAGE_SIZE, sort: 'name' }
  return useQuery({
    queryKey: dataSourceKeys.list(filters),
    queryFn: () => fetchDataSources(filters),
    select: (page) => page.items,
  })
}
