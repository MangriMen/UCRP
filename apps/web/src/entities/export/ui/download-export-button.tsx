import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDownloadExport } from '../api/use-download-export'

interface DownloadExportButtonProps {
  exportId: string
}

export function DownloadExportButton({ exportId }: DownloadExportButtonProps) {
  const download = useDownloadExport()
  return (
    <div className="flex flex-col gap-2">
      <Button
        variant="outline"
        size="sm"
        disabled={download.isPending}
        onClick={(event) => {
          event.stopPropagation()
          download.mutate(exportId)
        }}
      >
        Скачать CSV
      </Button>
      {download.error && <ErrorBanner error={download.error} />}
    </div>
  )
}
