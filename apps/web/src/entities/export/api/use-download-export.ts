import { useMutation } from '@tanstack/react-query'
import { requestBlob } from '@/shared/api/client'
import { saveBlob } from '@/shared/lib/save-blob'
import type { ExportFileFormat } from '../model/status'

interface DownloadExportVariables {
  exportId: string
  format: ExportFileFormat
}

export function useDownloadExport() {
  return useMutation({
    mutationFn: async ({ exportId, format }: DownloadExportVariables) => {
      const blob = await requestBlob(`/exports/${encodeURIComponent(exportId)}/download`)
      saveBlob(blob, `export-${exportId}.${format}`)
    },
  })
}
