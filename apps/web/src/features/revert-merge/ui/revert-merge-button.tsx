import { useState } from 'react'
import { MergeErrorAlert } from '@/entities/merge-candidate'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { useRevertMerge } from '../api/use-revert-merge'

interface RevertMergeButtonProps {
  candidateId: string
}

export function RevertMergeButton({ candidateId }: RevertMergeButtonProps) {
  const [open, setOpen] = useState(false)
  const mutation = useRevertMerge()

  return (
    <div className="flex flex-col gap-2">
      <Button variant="outline" onClick={() => setOpen(true)}>
        Разъединить
      </Button>
      {mutation.error && <MergeErrorAlert error={mutation.error} />}
      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title="Разъединить профили"
        description="События вернутся исходному профилю, для пары появится запрет объединения."
        confirmLabel="Разъединить"
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
