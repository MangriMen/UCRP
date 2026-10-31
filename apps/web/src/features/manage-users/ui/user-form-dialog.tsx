import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm, useWatch } from 'react-hook-form'
import { roleLabels, type User } from '@/entities/user'
import { useTenants } from '@/entities/tenant'
import { hasErrorCode } from '@/shared/api/errors'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { useCreateUser, useUpdateUser } from '../api/use-user-mutations'
import { assignableRoles, userFormSchema, type UserFormValues } from '../model/schema'

const TENANT_OPTIONS_LIMIT = 200

const roleItems = assignableRoles.map((role) => ({ value: role, label: roleLabels[role] }))

interface UserFormDialogProps {
  open: boolean
  user: User | null
  onOpenChange: (open: boolean) => void
}

export function UserFormDialog({ open, user, onOpenChange }: UserFormDialogProps) {
  const tenants = useTenants({ pageSize: TENANT_OPTIONS_LIMIT })
  const create = useCreateUser()
  const update = useUpdateUser()
  const mutation = user ? update : create
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<UserFormValues>({
    resolver: zodResolver(userFormSchema),
    defaultValues: {
      email: user?.email ?? '',
      displayName: user?.displayName ?? '',
      role: user && user.role !== 'system' ? user.role : 'marketer',
      tenantId: user?.tenantId ?? '',
      isActive: user?.isActive ?? true,
    },
  })
  const role = useWatch({ control, name: 'role' })

  const submit = handleSubmit((values) => {
    const tenantId = values.role === 'tenant' ? values.tenantId : null
    const options = {
      onSuccess: () => onOpenChange(false),
      onError: (error: unknown) => {
        if (hasErrorCode(error, 'duplicate')) {
          setError('email', { message: 'Пользователь с таким email уже существует' })
        }
      },
    }
    if (user) {
      const { displayName, role, isActive } = values
      update.mutate({ id: user.id, body: { displayName, role, tenantId, isActive } }, options)
    } else {
      create.mutate({ ...values, tenantId }, options)
    }
  })

  const tenantItems = (tenants.data?.items ?? []).map((t) => ({ value: t.id, label: t.name }))

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={user ? 'Изменить пользователя' : 'Новый пользователь'}
      submitLabel={user ? 'Сохранить' : 'Создать'}
      isPending={mutation.isPending}
      onSubmit={submit}
    >
      {mutation.error && !hasErrorCode(mutation.error, 'duplicate') && (
        <ErrorBanner error={mutation.error} />
      )}
      <FormField label="Email" error={errors.email?.message}>
        {(field) => (
          <Input {...field} type="email" disabled={user !== null} {...register('email')} />
        )}
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
                onValueChange={(next) => onChange(next)}
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
      <Controller
        control={control}
        name="isActive"
        render={({ field: { value, onChange } }) => (
          <CheckboxField label="Активен" checked={value} onCheckedChange={onChange} />
        )}
      />
    </FormDialog>
  )
}
