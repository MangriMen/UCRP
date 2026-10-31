import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { roleLabels } from '@/entities/user'
import { useTenants } from '@/entities/tenant'
import { hasErrorCode } from '@/shared/api/errors'
import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { useInviteUser } from '../api/use-invite-user'
import { invitableRoles, inviteFormSchema, type InviteFormValues } from '../model/schema'

const TENANT_OPTIONS_LIMIT = 200

const roleItems = invitableRoles.map((role) => ({ value: role, label: roleLabels[role] }))

interface InviteUserDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function InviteUserDialog({ open, onOpenChange }: InviteUserDialogProps) {
  const tenants = useTenants({ isActive: true, pageSize: TENANT_OPTIONS_LIMIT })
  const invite = useInviteUser()
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<InviteFormValues>({
    resolver: zodResolver(inviteFormSchema),
    defaultValues: { email: '', displayName: '', role: 'manager', tenantId: '' },
  })
  const role = useWatch({ control, name: 'role' })

  const submit = handleSubmit((values) => {
    const tenantId = values.role === 'tenant' ? values.tenantId : undefined
    invite.mutate(
      { email: values.email, displayName: values.displayName, role: values.role, tenantId },
      {
        onError: (error) => {
          if (hasErrorCode(error, 'conflict')) {
            setError('email', { message: 'Пользователь с таким email уже есть' })
          }
        },
      },
    )
  })

  if (invite.isSuccess) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Приглашение отправлено</DialogTitle>
            <DialogDescription>
              {`Ссылка для входа отправлена на ${invite.variables.email}. Она действует 24 часа и срабатывает один раз.`}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button onClick={() => onOpenChange(false)}>Готово</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    )
  }

  const tenantItems = (tenants.data?.items ?? []).map((t) => ({ value: t.id, label: t.name }))

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Пригласить пользователя"
      submitLabel="Отправить приглашение"
      isPending={invite.isPending}
      onSubmit={submit}
    >
      {invite.error && !hasErrorCode(invite.error, 'conflict') && (
        <ErrorBanner error={invite.error} />
      )}
      <FormField label="Email" error={errors.email?.message}>
        {(field) => <Input {...field} type="email" {...register('email')} />}
      </FormField>
      <FormField label="Имя" error={errors.displayName?.message}>
        {(field) => <Input {...field} {...register('displayName')} />}
      </FormField>
      <FormField label="Роль" error={errors.role?.message}>
        {(field) => (
          <Controller
            control={control}
            name="role"
            render={({ field: { value, onChange } }) => (
              <OptionSelect
                id={field.id}
                items={roleItems}
                value={value}
                onValueChange={onChange}
              />
            )}
          />
        )}
      </FormField>
      {role === 'tenant' && (
        <FormField label="Арендатор" error={errors.tenantId?.message}>
          {(field) => (
            <Controller
              control={control}
              name="tenantId"
              render={({ field: { value, onChange } }) => (
                <OptionSelect
                  id={field.id}
                  invalid={field['aria-invalid']}
                  items={tenantItems}
                  value={value}
                  onValueChange={onChange}
                />
              )}
            />
          )}
        </FormField>
      )}
    </FormDialog>
  )
}
