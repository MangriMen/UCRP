import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api } from '@/shared/api/api'
import { IMPORT_POLL_MS, pollWhile } from '@/shared/api/polling'
import { importKeys, type ImportListFilters } from '../model/keys'
import { isImportActive } from '../model/status'
import type { ImportBatch } from '../model/types'

export function useImports(filters: ImportListFilters = {}) {
  return useQuery({
    queryKey: importKeys.list(filters),
    queryFn: () =>
      api.get('/imports', {
        query: {
          data_source_id: filters.dataSourceId,
          status: filters.status,
          page: filters.page,
          page_size: filters.pageSize,
        },
      }),
    placeholderData: keepPreviousData,
    refetchInterval: pollWhile(
      (data) => data.items.some((batch) => isImportActive(batch.status)),
      IMPORT_POLL_MS,
    ),
  })
}

export function useImport(id: string) {
  return useQuery<ImportBatch>({
    queryKey: importKeys.detail(id),
    queryFn: () => api.get('/imports/{id}', { params: { id } }),
    refetchInterval: pollWhile((batch) => isImportActive(batch.status), IMPORT_POLL_MS),
  })
}
