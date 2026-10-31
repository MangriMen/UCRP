import { zodResolver } from '@hookform/resolvers/zod'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import type { Segment, SegmentTemplate } from '@/entities/segment'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { Textarea } from '@/shared/ui/textarea'
import { useCreateSegment, useUpdateSegment } from '../api/use-segment-mutations'
import {
  emptyGroupValues,
  MAX_GROUP_DEPTH,
  ruleFormSchema,
  segmentFormSchema,
  toRule,
  toRuleFormValues,
  type SegmentFormValues,
} from '../model/schema'
import { GroupEditor } from './group-editor'
import { PreviewPanel } from './preview-panel'
import { TemplateGallery } from './template-gallery'

interface RuleBuilderFormProps {
  segment?: Segment
  onDone: () => void
}

export function RuleBuilderForm({ segment, onDone }: RuleBuilderFormProps) {
  const create = useCreateSegment()
  const update = useUpdateSegment()
  const mutation = segment ? update : create
  const methods = useForm<SegmentFormValues>({
    resolver: zodResolver(segmentFormSchema),
    defaultValues: {
      ...(segment?.rule ? toRuleFormValues(segment.rule) : emptyGroupValues()),
      name: segment?.name ?? '',
      description: segment?.description ?? '',
      templateCode: null,
    },
  })

  const { register, control, handleSubmit, getValues, reset, formState } = methods
  const { errors } = formState
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

  return (
    <FormProvider {...methods}>
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
        <GroupEditor path="" remainingDepth={MAX_GROUP_DEPTH} />
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
    </FormProvider>
  )
}
