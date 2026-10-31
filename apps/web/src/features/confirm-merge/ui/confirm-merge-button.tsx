import { useState } from 'react'
import { MergeErrorAlert } from '@/entities/merge-candidate'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { useConfirmMerge } from '../api/use-confirm-merge'

interface ConfirmMergeButtonProps {
  candidateId: string
}

export function ConfirmMergeButton({ candidateId }: ConfirmMergeButtonProps) {
  const [open, setOpen] = useState(false)
  const mutation = useConfirmMerge()

  return (
    <div className="flex flex-col gap-2">
      <Button variant="default" onClick={() => setOpen(true)}>
        Объединить
      </Button>
      {mutation.error && <MergeErrorAlert error={mutation.error} />}
      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title="Объединить профили"
        description="Профили будут объединены, события и идентификаторы перейдут к основному профилю."
        confirmLabel="Объединить"
        destructive={false}
        isPending={mutation.isPending}
        onConfirm={() => {
          setOpen(false)
          mutation.mutate(candidateId)
        }}
      />
    </div>
  )
}
