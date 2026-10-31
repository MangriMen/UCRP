import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import type { DataSource } from '@/entities/data-source'
import { useDataSourceTypes } from '@/entities/data-source-type'
import { useTenants } from '@/entities/tenant'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { Textarea } from '@/shared/ui/textarea'
import { useCreateDataSource, useUpdateDataSource } from '../api/use-data-source-mutations'

const SHARED_SOURCE = 'shared'

const parseJsonObject = (text: string): Record<string, unknown> | null => {
  try {
    const value: unknown = JSON.parse(text)
    return typeof value === 'object' && value !== null && !Array.isArray(value)
      ? (value as Record<string, unknown>)
      : null
  } catch {
    return null
  }
}

const jsonObject = z
  .string()
  .refine((text) => text.trim() === '' || parseJsonObject(text) !== null, 'Введите JSON-объект')

const schema = z.object({
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  tenantId: z.string(),
  typeId: z.string().min(1, 'Выберите тип источника'),
  scheduleCron: z.string().trim(),
  fieldMapping: jsonObject,
  config: jsonObject,
  isActive: z.boolean(),
})

type FormValues = z.infer<typeof schema>

interface DataSourceFormDialogProps {
  open: boolean
  source: DataSource | null
  onOpenChange: (open: boolean) => void
}

const stringify = (value: object | null | undefined) =>
  value ? JSON.stringify(value, null, 2) : ''

export function DataSourceFormDialog({ open, source, onOpenChange }: DataSourceFormDialogProps) {
  const types = useDataSourceTypes()
  const tenants = useTenants({ pageSize: 200 })
  const create = useCreateDataSource()
  const update = useUpdateDataSource()
  const mutation = source ? update : create
  const {
    register,
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: source?.name ?? '',
      tenantId: source?.tenantId ?? SHARED_SOURCE,
      typeId: source?.typeId ?? '',
      scheduleCron: source?.scheduleCron ?? '',
      fieldMapping: stringify(source?.fieldMapping) || '{}',
      config: stringify(source?.config),
      isActive: source?.isActive ?? true,
    },
  })

  const submit = handleSubmit((values) => {
    const common = {
      name: values.name,
      scheduleCron: values.scheduleCron || null,
      fieldMapping: (parseJsonObject(values.fieldMapping) ?? {}) as Record<string, string>,
      config: values.config.trim() ? parseJsonObject(values.config) : null,
      isActive: values.isActive,
    }
    const options = { onSuccess: () => onOpenChange(false) }
    if (source) {
      update.mutate({ id: source.id, body: common }, options)
      return
    }
    create.mutate(
      {
        ...common,
        typeId: values.typeId,
        tenantId: values.tenantId === SHARED_SOURCE ? null : values.tenantId,
      },
      options,
    )
  })

  const typeItems = (types.data ?? []).map((t) => ({ value: t.id, label: t.name }))
  const tenantItems = [
    { value: SHARED_SOURCE, label: 'Общий источник' },
    ...(tenants.data?.items ?? []).map((t) => ({ value: t.id, label: t.name })),
  ]

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={source ? 'Изменить источник' : 'Новый источник'}
      submitLabel={source ? 'Сохранить' : 'Создать'}
      isPending={mutation.isPending}
      onSubmit={submit}
    >
      {mutation.error && <ErrorBanner error={mutation.error} />}
      <FormField label="Название" error={errors.name?.message}>
        {(field) => <Input {...field} {...register('name')} />}
      </FormField>
      <FormField label="Арендатор">
        {(field) => (
          <Controller
            control={control}
            name="tenantId"
            render={({ field: { value, onChange } }) => (
              <OptionSelect
                id={field.id}
                items={tenantItems}
                value={value}
                onValueChange={onChange}
                className={source ? 'pointer-events-none opacity-60' : undefined}
              />
            )}
          />
        )}
      </FormField>
      <FormField label="Тип источника" error={errors.typeId?.message}>
        {(field) => (
          <Controller
            control={control}
            name="typeId"
            render={({ field: { value, onChange } }) => (
              <OptionSelect
                id={field.id}
                invalid={field['aria-invalid']}
                items={typeItems}
                value={value}
                onValueChange={onChange}
                className={source ? 'pointer-events-none opacity-60' : undefined}
              />
            )}
          />
        )}
      </FormField>
      <FormField label="Расписание (cron)" error={errors.scheduleCron?.message}>
        {(field) => <Input {...field} placeholder="0 * * * *" {...register('scheduleCron')} />}
      </FormField>
      <FormField label="Сопоставление полей (JSON)" error={errors.fieldMapping?.message}>
        {(field) => <Textarea {...field} className="font-mono" {...register('fieldMapping')} />}
      </FormField>
      <FormField label="Конфигурация (JSON)" error={errors.config?.message}>
        {(field) => <Textarea {...field} className="font-mono" {...register('config')} />}
      </FormField>
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
