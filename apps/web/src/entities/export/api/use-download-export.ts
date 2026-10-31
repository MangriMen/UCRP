import { useMutation } from '@tanstack/react-query'
import { requestBlob } from '@/shared/api/client'
import { saveBlob } from '@/shared/lib/save-blob'

export function useDownloadExport() {
  return useMutation({
    mutationFn: async (exportId: string) => {
      const blob = await requestBlob(`/exports/${encodeURIComponent(exportId)}/download`)
      saveBlob(blob, `export-${exportId}.csv`)
    },
  })
}
