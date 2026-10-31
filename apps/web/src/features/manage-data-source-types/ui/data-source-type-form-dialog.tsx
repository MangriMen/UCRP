import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import type { DataSourceType } from '@/entities/data-source-type'
import { hasErrorCode } from '@/shared/api/errors'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import {
  useCreateDataSourceType,
  useUpdateDataSourceType,
} from '../api/use-data-source-type-mutations'

const schema = z.object({
  code: z.string().trim().min(1, 'Введите код').max(64, 'Не более 64 символов'),
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  connectorKey: z.string().trim().min(1, 'Введите ключ коннектора').max(64, 'Не более 64 символов'),
  supportsFile: z.boolean(),
  supportsPull: z.boolean(),
  isActive: z.boolean(),
})

type FormValues = z.infer<typeof schema>

interface DataSourceTypeFormDialogProps {
  open: boolean
  type: DataSourceType | null
  onOpenChange: (open: boolean) => void
}

export function DataSourceTypeFormDialog({
  open,
  type,
  onOpenChange,
}: DataSourceTypeFormDialogProps) {
  const create = useCreateDataSourceType()
  const update = useUpdateDataSourceType()
  const mutation = type ? update : create
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      code: type?.code ?? '',
      name: type?.name ?? '',
      connectorKey: type?.connectorKey ?? '',
      supportsFile: type?.supportsFile ?? false,
      supportsPull: type?.supportsPull ?? false,
      isActive: type?.isActive ?? true,
    },
  })

  const submit = handleSubmit((body) => {
    const options = {
      onSuccess: () => onOpenChange(false),
      onError: (error: unknown) => {
        if (hasErrorCode(error, 'duplicate')) {
          setError('code', { message: 'Тип источника с таким кодом уже существует' })
        }
      },
    }
    if (type) update.mutate({ id: type.id, body }, options)
    else create.mutate(body, options)
  })

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={type ? 'Изменить тип источника' : 'Новый тип источника'}
      submitLabel={type ? 'Сохранить' : 'Создать'}
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
      <FormField label="Ключ коннектора" error={errors.connectorKey?.message}>
        {(field) => <Input {...field} {...register('connectorKey')} />}
      </FormField>
      <Controller
        control={control}
        name="supportsFile"
        render={({ field: { value, onChange } }) => (
          <CheckboxField label="Загрузка файлов" checked={value} onCheckedChange={onChange} />
        )}
      />
      <Controller
        control={control}
        name="supportsPull"
        render={({ field: { value, onChange } }) => (
          <CheckboxField label="Получение по API" checked={value} onCheckedChange={onChange} />
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
