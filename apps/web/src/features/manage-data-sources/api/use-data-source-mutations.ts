import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  dataSourceKeys,
  type DataSourceCreateRequest,
  type DataSourceUpdateRequest,
} from '@/entities/data-source'
import { api } from '@/shared/api/api'

function useInvalidateDataSources() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: dataSourceKeys.all })
}

export function useCreateDataSource() {
  const invalidate = useInvalidateDataSources()
  return useMutation({
    mutationFn: (body: DataSourceCreateRequest) => api.post('/data-sources', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateDataSource() {
  const invalidate = useInvalidateDataSources()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: DataSourceUpdateRequest }) =>
      api.patch('/data-sources/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}

export function useDeleteDataSource() {
  const invalidate = useInvalidateDataSources()
  return useMutation({
    mutationFn: (id: string) => api.delete('/data-sources/{id}', { params: { id } }),
    onSuccess: invalidate,
  })
}
