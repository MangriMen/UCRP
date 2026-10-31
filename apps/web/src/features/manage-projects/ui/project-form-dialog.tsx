import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import { z } from 'zod'
import type { Project } from '@/entities/project'
import { useTenants } from '@/entities/tenant'
import { hasErrorCode } from '@/shared/api/errors'
import { CheckboxField } from '@/shared/ui/checkbox-field'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormDialog } from '@/shared/ui/form-dialog'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { useCreateProject, useUpdateProject } from '../api/use-project-mutations'

const TENANT_OPTIONS_LIMIT = 200

const schema = z.object({
  tenantId: z.string().min(1, 'Выберите арендатора'),
  code: z.string().trim().min(1, 'Введите код').max(64, 'Не более 64 символов'),
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  isActive: z.boolean(),
})

type FormValues = z.infer<typeof schema>

interface ProjectFormDialogProps {
  open: boolean
  project: Project | null
  onOpenChange: (open: boolean) => void
}

export function ProjectFormDialog({ open, project, onOpenChange }: ProjectFormDialogProps) {
  const tenants = useTenants({ pageSize: TENANT_OPTIONS_LIMIT })
  const create = useCreateProject()
  const update = useUpdateProject()
  const mutation = project ? update : create
  const {
    register,
    control,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      tenantId: project?.tenantId ?? '',
      code: project?.code ?? '',
      name: project?.name ?? '',
      isActive: project?.isActive ?? true,
    },
  })

  const submit = handleSubmit((body) => {
    const options = {
      onSuccess: () => onOpenChange(false),
      onError: (error: unknown) => {
        if (hasErrorCode(error, 'duplicate')) {
          setError('code', { message: 'Проект с таким кодом у арендатора уже существует' })
        }
      },
    }
    if (project) update.mutate({ id: project.id, body }, options)
    else create.mutate(body, options)
  })

  const tenantItems = (tenants.data?.items ?? []).map((t) => ({ value: t.id, label: t.name }))

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={project ? 'Изменить проект' : 'Новый проект'}
      submitLabel={project ? 'Сохранить' : 'Создать'}
      isPending={mutation.isPending}
      onSubmit={submit}
    >
      {mutation.error && !hasErrorCode(mutation.error, 'duplicate') && (
        <ErrorBanner error={mutation.error} />
      )}
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
      <FormField label="Код" error={errors.code?.message}>
        {(field) => <Input {...field} {...register('code')} />}
      </FormField>
      <FormField label="Название" error={errors.name?.message}>
        {(field) => <Input {...field} {...register('name')} />}
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
