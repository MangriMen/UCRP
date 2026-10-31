import { useState, type FormEvent } from 'react'
import { grantTransitionLabels, type AccessGrant, type GrantTransition } from '@/entities/access'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Textarea } from '@/shared/ui/textarea'
import { useGrantTransition } from '../api/use-grant-mutations'
import { GrantError } from './grant-error'

const descriptions: Record<GrantTransition, string> = {
  activate: 'Права гранта начнут действовать сразу после подтверждения.',
  suspend: 'Права перестанут действовать. Грант можно будет активировать снова.',
  revoke:
    'Отзыв необратим: грант больше нельзя будет активировать или изменить. Чтобы вернуть доступ, выдайте новый грант.',
}

interface GrantTransitionDialogProps {
  grant: AccessGrant
  transition: GrantTransition
  onClose: () => void
}

export function GrantTransitionDialog({ grant, transition, onClose }: GrantTransitionDialogProps) {
  const mutation = useGrantTransition()
  const [reason, setReason] = useState('')
  const [error, setError] = useState<string>()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!reason.trim()) {
      setError('Укажите причину')
      return
    }
    mutation.mutate({ id: grant.id, transition, reason: reason.trim() }, { onSuccess: onClose })
  }

  return (
    <FormDialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={`${grantTransitionLabels[transition]} грант?`}
      submitLabel={grantTransitionLabels[transition]}
      isPending={mutation.isPending}
      onSubmit={submit}
    >
      <p className="text-sm text-muted-foreground">{descriptions[transition]}</p>
      {mutation.error && <GrantError error={mutation.error} />}
      <FormField label="Причина" error={error}>
        {(field) => (
          <Textarea {...field} value={reason} onChange={(event) => setReason(event.target.value)} />
        )}
      </FormField>
    </FormDialog>
  )
}
