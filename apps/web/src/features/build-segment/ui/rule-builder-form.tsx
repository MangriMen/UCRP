import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm, useWatch } from 'react-hook-form'
import type { Segment, SegmentTemplate } from '@/entities/segment'
import { useTenantCategories } from '@/entities/tenant-category'
import { useTenants } from '@/entities/tenant'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { Textarea } from '@/shared/ui/textarea'
import { useCreateSegment, useUpdateSegment } from '../api/use-segment-mutations'
import {
  emptyRuleValues,
  ruleFormSchema,
  segmentFormSchema,
  toRule,
  toRuleFormValues,
  type SegmentFormValues,
} from '../model/schema'
import { EntityChecklist } from './entity-checklist'
import { PreviewPanel } from './preview-panel'
import { TemplateGallery } from './template-gallery'

const REFERENCE_LIMIT = 200

const contactItems = [
  { value: 'none', label: 'Не требуется' },
  { value: 'phone', label: 'Телефон' },
  { value: 'email', label: 'Email' },
]

interface RuleBuilderFormProps {
  segment?: Segment
  onDone: () => void
}

export function RuleBuilderForm({ segment, onDone }: RuleBuilderFormProps) {
  const tenants = useTenants({ pageSize: REFERENCE_LIMIT })
  const categories = useTenantCategories()
  const create = useCreateSegment()
  const update = useUpdateSegment()
  const mutation = segment ? update : create
  const {
    register,
    control,
    handleSubmit,
    getValues,
    reset,
    formState: { errors },
  } = useForm<SegmentFormValues>({
    resolver: zodResolver(segmentFormSchema),
    defaultValues: {
      ...(segment?.rule ? toRuleFormValues(segment.rule) : emptyRuleValues),
      name: segment?.name ?? '',
      description: segment?.description ?? '',
      templateCode: null,
    },
  })

  const watched = useWatch({ control })
  const parsedRule = ruleFormSchema.safeParse(watched)
  const previewRule = parsedRule.success ? toRule(parsedRule.data) : null
  const templateCode = watched.templateCode ?? null

  const applyTemplate = (template: SegmentTemplate) => {
    const { name, description } = getValues()
    reset({
      ...toRuleFormValues(template.rule),
      name: name || template.name,
      description: description || template.description || '',
      templateCode: template.code,
    })
  }

  const tenantItems = (tenants.data?.items ?? []).map((t) => ({ value: t.id, label: t.name }))
  const categoryItems = (categories.data ?? []).map((c) => ({ value: c.id, label: c.name }))

  const submit = handleSubmit((values) => {
    const rule = toRule(values)
    const name = values.name
    const description = values.description || null
    if (segment) {
      update.mutate({ id: segment.id, body: { name, description, rule } }, { onSuccess: onDone })
    } else {
      create.mutate(
        { name, description, templateCode: values.templateCode, rule },
        { onSuccess: onDone },
      )
    }
  })

  const checklist = (
    name: 'was' | 'wasNot' | 'categoriesWas' | 'categoriesWasNot',
    label: string,
    items: typeof tenantItems,
    hint: string,
  ) => (
    <FormField label={label} hint={hint} error={errors[name]?.message}>
      {() => (
        <Controller
          control={control}
          name={name}
          render={({ field: { value, onChange } }) => (
            <EntityChecklist items={items} value={value} onChange={onChange} />
          )}
        />
      )}
    </FormField>
  )

  return (
    <form onSubmit={submit} noValidate className="flex flex-col gap-4">
      {mutation.error && <ErrorBanner error={mutation.error} />}
      {!segment && <TemplateGallery selectedCode={templateCode} onSelect={applyTemplate} />}
      <FormField label="Название" error={errors.name?.message}>
        {(field) => <Input {...field} {...register('name')} />}
      </FormField>
      <FormField label="Описание" error={errors.description?.message}>
        {(field) => <Textarea {...field} {...register('description')} />}
      </FormField>
      <p className="text-sm font-medium">Условия</p>
      <p className="-mt-2 text-xs text-muted-foreground">
        Все заполненные условия должны выполняться одновременно, внутри списка достаточно любого
        варианта.
      </p>
      <FormField label="Период анализа, дней" error={errors.periodDays?.message}>
        {(field) => <Input {...field} inputMode="numeric" {...register('periodDays')} />}
      </FormField>
      {checklist('was', 'Был у арендаторов', tenantItems, 'Хотя бы у одного из выбранных')}
      {checklist('wasNot', 'Не был у арендаторов', tenantItems, 'Ни у одного из выбранных')}
      {checklist(
        'categoriesWas',
        'Был в категориях',
        categoryItems,
        'Хотя бы в одной из выбранных',
      )}
      {checklist(
        'categoriesWasNot',
        'Не был в категориях',
        categoryItems,
        'Ни в одной из выбранных',
      )}
      <FormField label="Контакт гостя">
        {(field) => (
          <Controller
            control={control}
            name="requireContact"
            render={({ field: { value, onChange } }) => (
              <OptionSelect
                id={field.id}
                items={contactItems}
                value={value}
                onValueChange={(next) => onChange(next)}
              />
            )}
          />
        )}
      </FormField>
      <FormField label="Минимум визитов" error={errors.minVisits?.message}>
        {(field) => <Input {...field} inputMode="numeric" {...register('minVisits')} />}
      </FormField>
      <FormField label="Минимальная сумма, ₽" error={errors.minAmount?.message}>
        {(field) => <Input {...field} inputMode="decimal" {...register('minAmount')} />}
      </FormField>
      <PreviewPanel rule={previewRule} />
      <DialogFooter>
        <Button type="button" variant="outline" onClick={onDone}>
          Отмена
        </Button>
        <Button type="submit" disabled={mutation.isPending}>
          {segment ? 'Сохранить новую версию' : 'Создать сегмент'}
        </Button>
      </DialogFooter>
    </form>
  )
}
