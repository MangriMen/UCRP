import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDownloadExport } from '../api/use-download-export'
import { exportFileFormat, exportFileFormatLabels } from '../model/status'
import type { ExportFormat } from '../model/types'

interface DownloadExportButtonProps {
  exportId: string
  format: ExportFormat
}

export function DownloadExportButton({ exportId, format }: DownloadExportButtonProps) {
  const download = useDownloadExport()
  const fileFormat = exportFileFormat(format)
  return (
    <div className="flex flex-col gap-2">
      <Button
        variant="outline"
        size="sm"
        disabled={download.isPending}
        onClick={(event) => {
          event.stopPropagation()
          download.mutate({ exportId, format: fileFormat })
        }}
      >
        Скачать {exportFileFormatLabels[fileFormat]}
      </Button>
      {download.error && <ErrorBanner error={download.error} />}
    </div>
  )
}
