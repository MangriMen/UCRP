import { useMutation, useQueryClient } from '@tanstack/react-query'
import { dataSourceTypeKeys, type DataSourceTypeRequest } from '@/entities/data-source-type'
import { api } from '@/shared/api/api'

function useInvalidateDataSourceTypes() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: dataSourceTypeKeys.all })
}

export function useCreateDataSourceType() {
  const invalidate = useInvalidateDataSourceTypes()
  return useMutation({
    mutationFn: (body: DataSourceTypeRequest) => api.post('/data-source-types', { body }),
    onSuccess: invalidate,
  })
}

export function useUpdateDataSourceType() {
  const invalidate = useInvalidateDataSourceTypes()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: DataSourceTypeRequest }) =>
      api.patch('/data-source-types/{id}', { params: { id }, body }),
    onSuccess: invalidate,
  })
}
