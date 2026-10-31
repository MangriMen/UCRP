import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import type { TenantCategory } from '@/entities/tenant-category'
import { hasErrorCode } from '@/shared/api/errors'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { Textarea } from '@/shared/ui/textarea'
import {
  useCreateTenantCategory,
  useUpdateTenantCategory,
} from '../api/use-tenant-category-mutations'

const schema = z.object({
  code: z.string().trim().min(1, 'Введите код').max(64, 'Не более 64 символов'),
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  description: z.string().trim().max(1000, 'Не более 1000 символов'),
})

type FormValues = z.infer<typeof schema>

interface TenantCategoryFormDialogProps {
  open: boolean
  category: TenantCategory | null
  onOpenChange: (open: boolean) => void
}

export function TenantCategoryFormDialog({
  open,
  category,
  onOpenChange,
}: TenantCategoryFormDialogProps) {
  const create = useCreateTenantCategory()
  const update = useUpdateTenantCategory()
  const mutation = category ? update : create
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      code: category?.code ?? '',
      name: category?.name ?? '',
      description: category?.description ?? '',
    },
  })

  const submit = handleSubmit((values) => {
    const body = { ...values, description: values.description || null }
    const options = {
      onSuccess: () => onOpenChange(false),
      onError: (error: unknown) => {
        if (hasErrorCode(error, 'duplicate')) {
          setError('code', { message: 'Категория с таким кодом уже существует' })
        }
      },
    }
    if (category) update.mutate({ id: category.id, body }, options)
    else create.mutate(body, options)
  })

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={category ? 'Изменить категорию' : 'Новая категория'}
      submitLabel={category ? 'Сохранить' : 'Создать'}
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
      <FormField label="Описание" error={errors.description?.message}>
        {(field) => <Textarea {...field} {...register('description')} />}
      </FormField>
    </FormDialog>
  )
}
