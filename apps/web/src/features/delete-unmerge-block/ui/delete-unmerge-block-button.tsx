import { useState } from 'react'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteUnmergeBlock } from '../api/use-delete-unmerge-block'

export function DeleteUnmergeBlockButton({ blockId }: { blockId: string }) {
  const [open, setOpen] = useState(false)
  const mutation = useDeleteUnmergeBlock()

  return (
    <div className="flex flex-col gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={(event) => {
          event.stopPropagation()
          setOpen(true)
        }}
      >
        Снять запрет
      </Button>
      {mutation.error && <ErrorBanner error={mutation.error} />}
      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title="Снять запрет склейки"
        description="После снятия запрета пара снова может быть склеена автоматически."
        confirmLabel="Снять запрет"
        isPending={mutation.isPending}
        onConfirm={() => {
          setOpen(false)
          mutation.mutate(blockId)
        }}
      />
    </div>
  )
}
