import { useMutation, useQueryClient } from '@tanstack/react-query'
import { exportKeys, type ExportCreateRequest } from '@/entities/export'
import { api } from '@/shared/api/api'

interface CreateExportInput {
  request: ExportCreateRequest
  idempotencyKey: string
}

export function useCreateExport() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ request, idempotencyKey }: CreateExportInput) =>
      api.post('/exports', { body: request, headers: { 'Idempotency-Key': idempotencyKey } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: exportKeys.all }),
  })
}
