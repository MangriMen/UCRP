import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { projectKeys, type ProjectListFilters } from '../model/keys'

export function useProjects(filters: ProjectListFilters = {}) {
  return useQuery({
    queryKey: projectKeys.list(filters),
    queryFn: () =>
      api.get('/projects', {
        query: { tenant_id: filters.tenantId, page: filters.page, page_size: filters.pageSize },
      }),
    placeholderData: keepPreviousData,
  })
}
