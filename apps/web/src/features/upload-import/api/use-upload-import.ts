import { useMutation, useQueryClient } from '@tanstack/react-query'
import { importKeys } from '@/entities/import-batch'
import { api } from '@/shared/api/api'

interface UploadImportInput {
  dataSourceId: string
  file: File
}

export function useUploadImport() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ dataSourceId, file }: UploadImportInput) => {
      const body = new FormData()
      body.set('data_source_id', dataSourceId)
      body.set('file', file)
      return api.post('/imports', { body })
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: importKeys.all }),
  })
}
