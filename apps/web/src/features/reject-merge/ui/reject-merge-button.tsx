import { useState } from 'react'
import { MergeErrorAlert } from '@/entities/merge-candidate'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { useRejectMerge } from '../api/use-reject-merge'

interface RejectMergeButtonProps {
  candidateId: string
}

export function RejectMergeButton({ candidateId }: RejectMergeButtonProps) {
  const [open, setOpen] = useState(false)
  const mutation = useRejectMerge()

  return (
    <div className="flex flex-col gap-2">
      <Button variant="outline" onClick={() => setOpen(true)}>
        Отклонить
      </Button>
      {mutation.error && <MergeErrorAlert error={mutation.error} />}
      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title="Отклонить склейку"
        description="Пара будет отклонена, автоматически её больше не склеят."
        confirmLabel="Отклонить"
        destructive={true}
        isPending={mutation.isPending}
        onConfirm={() => {
          setOpen(false)
          mutation.mutate(candidateId)
        }}
      />
    </div>
  )
}
