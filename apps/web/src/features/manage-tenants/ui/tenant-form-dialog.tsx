import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import { useMalls } from '@/entities/mall'
import type { Tenant } from '@/entities/tenant'
import { useTenantCategories } from '@/entities/tenant-category'
import { hasErrorCode } from '@/shared/api/errors'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { useCreateTenant, useUpdateTenant } from '../api/use-tenant-mutations'

const schema = z.object({
  code: z.string().trim().min(1, 'Введите код').max(64, 'Не более 64 символов'),
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  categoryId: z.string().min(1, 'Выберите категорию'),
  mallId: z.string().min(1, 'Выберите ТЦ'),
  floor: z
    .string()
    .trim()
    .regex(/^(-?\d+)?$/, 'Введите целый номер этажа или оставьте пустым'),
  priority: z
    .string()
    .regex(/^\d+$/, 'Введите число от 0 до 100')
    .refine((v) => Number(v) <= 100, 'Введите число от 0 до 100'),
  isActive: z.boolean(),
})

type FormValues = z.infer<typeof schema>

const DUPLICATE_CODE_ERROR = 'conflict'

interface TenantFormDialogProps {
  open: boolean
  tenant: Tenant | null
  onOpenChange: (open: boolean) => void
}

export function TenantFormDialog({ open, tenant, onOpenChange }: TenantFormDialogProps) {
  const categories = useTenantCategories()
  const malls = useMalls()
  const create = useCreateTenant()
  const update = useUpdateTenant()
  const mutation = tenant ? update : create
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      code: tenant?.code ?? '',
      name: tenant?.name ?? '',
      categoryId: tenant?.categoryId ?? '',
      mallId: tenant?.mallId ?? '',
      floor: tenant?.floor != null ? String(tenant.floor) : '',
      priority: String(tenant?.priority ?? 50),
      isActive: tenant?.isActive ?? true,
    },
  })

  const submit = handleSubmit((values) => {
    const body = {
      code: values.code,
      name: values.name,
      categoryId: values.categoryId,
      mallId: values.mallId,
      floor: values.floor === '' ? null : Number(values.floor),
      priority: Number(values.priority),
      isActive: values.isActive,
    }
    const options = {
      onSuccess: () => onOpenChange(false),
      onError: (error: unknown) => {
        if (hasErrorCode(error, DUPLICATE_CODE_ERROR)) {
          setError('code', { message: 'Арендатор с таким кодом уже существует' })
        }
      },
    }
    if (tenant) update.mutate({ id: tenant.id, body }, options)
    else create.mutate(body, options)
  })

  const categoryItems = (categories.data ?? []).map((c) => ({ value: c.id, label: c.name }))
  const mallItems = (malls.data ?? []).map((m) => ({ value: m.id, label: m.name }))

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={tenant ? 'Изменить арендатора' : 'Новый арендатор'}
      submitLabel={tenant ? 'Сохранить' : 'Создать'}
      isPending={mutation.isPending}
      onSubmit={submit}
    >
      {mutation.error && !hasErrorCode(mutation.error, DUPLICATE_CODE_ERROR) && (
        <ErrorBanner error={mutation.error} />
      )}
      <FormField label="Код" error={errors.code?.message}>
        {(field) => <Input {...field} {...register('code')} />}
      </FormField>
      <FormField label="Название" error={errors.name?.message}>
        {(field) => <Input {...field} {...register('name')} />}
      </FormField>
      <FormField label="Категория" error={errors.categoryId?.message}>
        {(field) => (
          <Controller
            control={control}
            name="categoryId"
            render={({ field: { value, onChange } }) => (
              <OptionSelect
                id={field.id}
                invalid={field['aria-invalid']}
                items={categoryItems}
                value={value}
                onValueChange={onChange}
              />
            )}
          />
        )}
      </FormField>
      <FormField label="ТЦ" error={errors.mallId?.message}>
        {(field) => (
          <Controller
            control={control}
            name="mallId"
            render={({ field: { value, onChange } }) => (
              <OptionSelect
                id={field.id}
                invalid={field['aria-invalid']}
                items={mallItems}
                value={value}
                onValueChange={onChange}
              />
            )}
          />
        )}
      </FormField>
      <div className="grid grid-cols-2 gap-3">
        <FormField label="Этаж" hint="Пусто — без этажа" error={errors.floor?.message}>
          {(field) => <Input {...field} inputMode="numeric" {...register('floor')} />}
        </FormField>
        <FormField label="Приоритет" error={errors.priority?.message}>
          {(field) => <Input {...field} inputMode="numeric" {...register('priority')} />}
        </FormField>
      </div>
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
