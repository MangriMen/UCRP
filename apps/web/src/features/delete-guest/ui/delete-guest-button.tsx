import { useState } from 'react'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteGuest } from '../api/use-delete-guest'

interface DeleteGuestButtonProps {
  guestId: string
  onDeleted: () => void
}

export function DeleteGuestButton({ guestId, onDeleted }: DeleteGuestButtonProps) {
  const [open, setOpen] = useState(false)
  const mutation = useDeleteGuest()

  return (
    <>
      <Button variant="outline" onClick={() => setOpen(true)}>
        Удалить
      </Button>
      {mutation.error && <ErrorBanner error={mutation.error} />}
      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title="Удалить профиль гостя"
        description="Профиль будет скрыт из поиска. События и идентификаторы сохранятся для аудита."
        confirmLabel="Удалить"
        destructive
        isPending={mutation.isPending}
        onConfirm={() => {
          setOpen(false)
          mutation.mutate(guestId, { onSuccess: onDeleted })
        }}
      />
    </>
  )
}
