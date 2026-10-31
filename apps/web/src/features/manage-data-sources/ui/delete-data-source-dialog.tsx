import type { DataSource } from '@/entities/data-source'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteDataSource } from '../api/use-data-source-mutations'

interface DeleteDataSourceDialogProps {
  source: DataSource | null
  onClose: () => void
}

export function DeleteDataSourceDialog({ source, onClose }: DeleteDataSourceDialogProps) {
  const remove = useDeleteDataSource()

  return (
    <>
      {remove.error && <ErrorBanner error={remove.error} />}
      <ConfirmDialog
        open={source !== null}
        onOpenChange={(open) => !open && onClose()}
        title="Удалить источник?"
        description={`Источник «${source?.name ?? ''}» будет деактивирован. Если у него есть импорты, удаление будет отклонено.`}
        confirmLabel="Удалить"
        destructive
        isPending={remove.isPending}
        onConfirm={() => {
          if (source) remove.mutate(source.id, { onSettled: onClose })
        }}
      />
    </>
  )
}
