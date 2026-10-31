import { useState, type FormEvent } from 'react'
import { MergeErrorAlert, type MergeResult } from '@/entities/merge-candidate'
import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { useMergeGuests } from '../api/use-merge-guests'

interface MergeGuestsDialogProps {
  guestId: string
  onMerged: (result: MergeResult) => void
}

export function MergeGuestsDialog({ guestId, onMerged }: MergeGuestsDialogProps) {
  const [open, setOpen] = useState(false)
  const [targetGuestId, setTargetGuestId] = useState('')
  const mutation = useMergeGuests()
  const target = targetGuestId.trim()
  const isSelf = target === guestId

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!target || isSelf) return
    mutation.mutate(
      { guestId, targetGuestId: target },
      {
        onSuccess: (result) => {
          setOpen(false)
          onMerged(result)
        },
      },
    )
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) {
          setTargetGuestId('')
          mutation.reset()
        }
      }}
    >
      <DialogTrigger render={<Button variant="outline" />}>
        Объединить с другим профилем
      </DialogTrigger>
      <DialogContent>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>Ручное объединение</DialogTitle>
            <DialogDescription>
              Этот профиль будет объединён в выбранный: он станет основным.
            </DialogDescription>
          </DialogHeader>
          {mutation.error && <MergeErrorAlert error={mutation.error} />}
          <FormField
            label="ID основного профиля"
            error={isSelf ? 'Нельзя объединить профиль с самим собой' : undefined}
          >
            {(field) => (
              <Input
                {...field}
                value={targetGuestId}
                onChange={(event) => setTargetGuestId(event.target.value)}
              />
            )}
          </FormField>
          <DialogFooter>
            <Button type="submit" disabled={!target || isSelf || mutation.isPending}>
              Объединить
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
