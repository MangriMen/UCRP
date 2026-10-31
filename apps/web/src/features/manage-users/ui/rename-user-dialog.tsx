import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import type { User } from '@/entities/user'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { useUpdateUser } from '../api/use-user-mutations'
import { renameUserSchema, type RenameUserValues } from '../model/schema'

interface RenameUserDialogProps {
  user: User
  onOpenChange: (open: boolean) => void
}

export function RenameUserDialog({ user, onOpenChange }: RenameUserDialogProps) {
  const update = useUpdateUser()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RenameUserValues>({
    resolver: zodResolver(renameUserSchema),
    defaultValues: { displayName: user.displayName },
  })

  const submit = handleSubmit(({ displayName }) =>
    update.mutate({ id: user.id, body: { displayName } }, { onSuccess: () => onOpenChange(false) }),
  )

  return (
    <FormDialog
      open
      onOpenChange={onOpenChange}
      title="Изменить имя"
      submitLabel="Сохранить"
      isPending={update.isPending}
      onSubmit={submit}
    >
      {update.error && <ErrorBanner error={update.error} />}
      <FormField label="Email">
        {(field) => <Input {...field} type="email" value={user.email} disabled readOnly />}
      </FormField>
      <FormField label="Имя" error={errors.displayName?.message}>
        {(field) => <Input {...field} {...register('displayName')} />}
      </FormField>
    </FormDialog>
  )
}
