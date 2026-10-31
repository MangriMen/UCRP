import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import type { Mall } from '@/entities/mall'
import { hasErrorCode } from '@/shared/api/errors'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { useCreateMall, useUpdateMall } from '../api/use-mall-mutations'

const schema = z.object({
  code: z.string().trim().min(1, 'Введите код').max(64, 'Не более 64 символов'),
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  isActive: z.boolean(),
})

type FormValues = z.infer<typeof schema>

interface MallFormDialogProps {
  open: boolean
  mall: Mall | null
  onOpenChange: (open: boolean) => void
}

export function MallFormDialog({ open, mall, onOpenChange }: MallFormDialogProps) {
  const create = useCreateMall()
  const update = useUpdateMall()
  const mutation = mall ? update : create
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      code: mall?.code ?? '',
      name: mall?.name ?? '',
      isActive: mall?.isActive ?? true,
    },
  })

  const submit = handleSubmit((values) => {
    const options = {
      onSuccess: () => onOpenChange(false),
      onError: (error: unknown) => {
        if (hasErrorCode(error, 'duplicate')) {
          setError('code', { message: 'ТЦ с таким кодом уже существует' })
        }
      },
    }
    if (mall) {
      update.mutate(
        { id: mall.id, body: { name: values.name, isActive: values.isActive } },
        options,
      )
    } else {
      create.mutate({ code: values.code, name: values.name }, options)
    }
  })

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={mall ? 'Изменить ТЦ' : 'Новый ТЦ'}
      submitLabel={mall ? 'Сохранить' : 'Создать'}
      isPending={mutation.isPending}
      onSubmit={submit}
    >
      {mutation.error && !hasErrorCode(mutation.error, 'duplicate') && (
        <ErrorBanner error={mutation.error} />
      )}
      {!mall && (
        <FormField label="Код" error={errors.code?.message}>
          {(field) => <Input {...field} {...register('code')} />}
        </FormField>
      )}
      <FormField label="Название" error={errors.name?.message}>
        {(field) => <Input {...field} {...register('name')} />}
      </FormField>
      {mall && (
        <Controller
          control={control}
          name="isActive"
          render={({ field: { value, onChange } }) => (
            <CheckboxField label="Активен" checked={value} onCheckedChange={onChange} />
          )}
        />
      )}
    </FormDialog>
  )
}
