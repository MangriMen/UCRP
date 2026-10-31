import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import type { EventType } from '@/entities/event-type'
import { hasErrorCode } from '@/shared/api/errors'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { useCreateEventType, useUpdateEventType } from '../api/use-event-type-mutations'

const schema = z.object({
  code: z.string().trim().min(1, 'Введите код').max(64, 'Не более 64 символов'),
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  hasAmount: z.boolean(),
  countsAsPresence: z.boolean(),
  isActive: z.boolean(),
})

type FormValues = z.infer<typeof schema>

interface EventTypeFormDialogProps {
  open: boolean
  eventType: EventType | null
  onOpenChange: (open: boolean) => void
}

export function EventTypeFormDialog({ open, eventType, onOpenChange }: EventTypeFormDialogProps) {
  const create = useCreateEventType()
  const update = useUpdateEventType()
  const mutation = eventType ? update : create
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      code: eventType?.code ?? '',
      name: eventType?.name ?? '',
      hasAmount: eventType?.hasAmount ?? false,
      countsAsPresence: eventType?.countsAsPresence ?? false,
      isActive: eventType?.isActive ?? true,
    },
  })

  const submit = handleSubmit((body) => {
    const options = {
      onSuccess: () => onOpenChange(false),
      onError: (error: unknown) => {
        if (hasErrorCode(error, 'duplicate')) {
          setError('code', { message: 'Тип события с таким кодом уже существует' })
        }
      },
    }
    if (eventType) update.mutate({ id: eventType.id, body }, options)
    else create.mutate(body, options)
  })

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={eventType ? 'Изменить тип события' : 'Новый тип события'}
      submitLabel={eventType ? 'Сохранить' : 'Создать'}
      isPending={mutation.isPending}
      onSubmit={submit}
    >
      {mutation.error && !hasErrorCode(mutation.error, 'duplicate') && (
        <ErrorBanner error={mutation.error} />
      )}
      <FormField label="Код" error={errors.code?.message}>
        {(field) => <Input {...field} {...register('code')} />}
      </FormField>
      <FormField label="Название" error={errors.name?.message}>
        {(field) => <Input {...field} {...register('name')} />}
      </FormField>
      <Controller
        control={control}
        name="hasAmount"
        render={({ field: { value, onChange } }) => (
          <CheckboxField
            label="Событие содержит сумму"
            checked={value}
            onCheckedChange={onChange}
          />
        )}
      />
      <Controller
        control={control}
        name="countsAsPresence"
        render={({ field: { value, onChange } }) => (
          <CheckboxField label="Считается визитом" checked={value} onCheckedChange={onChange} />
        )}
      />
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
