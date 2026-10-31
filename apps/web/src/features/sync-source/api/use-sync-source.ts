import { useMutation, useQueryClient } from '@tanstack/react-query'
import { dataSourceKeys } from '@/entities/data-source'
import { importKeys } from '@/entities/import-batch'
import { api } from '@/shared/api/api'

export function useSyncSource() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (sourceId: string) =>
      api.post('/data-sources/{id}/sync', { params: { id: sourceId } }),
    onSettled: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: importKeys.lists() }),
        queryClient.invalidateQueries({ queryKey: dataSourceKeys.all }),
      ]),
  })
}
