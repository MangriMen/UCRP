import { Controller, get, useFormContext } from 'react-hook-form'
import { useMalls } from '@/entities/mall'
import { useTenantCategories } from '@/entities/tenant-category'
import { useTenants } from '@/entities/tenant'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { zoneKey, type SegmentFormValues } from '../model/schema'
import { EntityChecklist } from './entity-checklist'

const REFERENCE_LIMIT = 200

const contactItems = [
  { value: 'none', label: 'Не требуется' },
  { value: 'phone', label: 'Телефон' },
  { value: 'email', label: 'Email' },
]

type ChecklistField =
  | 'visitedTenantIds'
  | 'excludedTenantIds'
  | 'visitedCategoryIds'
  | 'excludedCategoryIds'
  | 'purchasedZones'

interface LeafFieldsProps {
  path: string
}

export function LeafFields({ path }: LeafFieldsProps) {
  const { register, control, formState } = useFormContext<SegmentFormValues>()
  const tenants = useTenants({ pageSize: REFERENCE_LIMIT })
  const categories = useTenantCategories()
  const malls = useMalls()
  const tenantItems = (tenants.data?.items ?? []).map((t) => ({ value: t.id, label: t.name }))
  const categoryItems = (categories.data ?? []).map((c) => ({ value: c.id, label: c.name }))
  const zoneItems = (malls.data ?? []).flatMap((mall) =>
    mall.floors.map((floor) => ({
      value: zoneKey({ mallId: mall.id, floor }),
      label: `${mall.name} · ${floor} этаж`,
    })),
  )
  const fieldName = (name: string) => `${path}.${name}` as 'conditions.0.periodDays'
  const errorOf = (name: string): string | undefined =>
    (get(formState.errors, `${path}.${name}`) as { message?: string } | undefined)?.message

  const checklist = (
    name: ChecklistField,
    label: string,
    items: typeof tenantItems,
    hint: string,
  ) => (
    <FormField label={label} hint={hint} error={errorOf(name)}>
      {() => (
        <Controller
          control={control}
          name={fieldName(name) as 'conditions.0.visitedTenantIds'}
          render={({ field: { value, onChange } }) => (
            <EntityChecklist items={items} value={value} onChange={onChange} />
          )}
        />
      )}
    </FormField>
  )

  return (
    <div className="flex flex-col gap-4">
      <FormField label="Период анализа, дней" error={errorOf('periodDays')}>
        {(field) => <Input {...field} inputMode="numeric" {...register(fieldName('periodDays'))} />}
      </FormField>
      {checklist(
        'visitedTenantIds',
        'Был у арендаторов',
        tenantItems,
        'Хотя бы у одного из выбранных',
      )}
      {checklist(
        'excludedTenantIds',
        'Не был у арендаторов',
        tenantItems,
        'Ни у одного из выбранных',
      )}
      {checklist(
        'purchasedZones',
        'Покупал на этажах',
        zoneItems,
        'Хотя бы одна покупка на любом из выбранных этажей',
      )}
      {checklist(
        'visitedCategoryIds',
        'Был в категориях',
        categoryItems,
        'Хотя бы в одной из выбранных',
      )}
      {checklist(
        'excludedCategoryIds',
        'Не был в категориях',
        categoryItems,
        'Ни в одной из выбранных',
      )}
      <FormField label="Контакт гостя">
        {(field) => (
          <Controller
            control={control}
            name={fieldName('requireContact') as 'conditions.0.requireContact'}
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
      <FormField label="Минимум визитов" error={errorOf('minVisits')}>
        {(field) => (
          <Input
            {...field}
            inputMode="numeric"
            {...register(fieldName('minVisits') as 'conditions.0.minVisits')}
          />
        )}
      </FormField>
      <FormField label="Минимальная сумма, ₽" error={errorOf('minAmount')}>
        {(field) => (
          <Input
            {...field}
            inputMode="decimal"
            {...register(fieldName('minAmount') as 'conditions.0.minAmount')}
          />
        )}
      </FormField>
    </div>
  )
}
