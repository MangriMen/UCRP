import { z } from 'zod'
import type { SegmentRule } from '@/entities/segment'

export const MAX_PERIOD_DAYS = 730

export const contactOptions = ['none', 'phone', 'email'] as const

const MINOR_PER_MAJOR = 100

const isIntegerInRange = (value: string, min: number, max: number) => {
  if (!/^\d+$/.test(value)) return false
  const number = Number(value)
  return number >= min && number <= max
}

const ruleFields = z.object({
  periodDays: z
    .string()
    .trim()
    .refine((value) => isIntegerInRange(value, 1, MAX_PERIOD_DAYS), {
      message: `Целое число от 1 до ${MAX_PERIOD_DAYS}`,
    }),
  was: z.array(z.string()),
  wasNot: z.array(z.string()),
  categoriesWas: z.array(z.string()),
  categoriesWasNot: z.array(z.string()),
  requireContact: z.enum(contactOptions),
  minVisits: z
    .string()
    .trim()
    .refine((value) => value === '' || isIntegerInRange(value, 1, Number.MAX_SAFE_INTEGER), {
      message: 'Целое число не меньше 1',
    }),
  minAmount: z
    .string()
    .trim()
    .refine((value) => value === '' || /^\d+([.,]\d{1,2})?$/.test(value), {
      message: 'Неотрицательная сумма, не более 2 знаков после запятой',
    }),
})

type RuleFields = z.infer<typeof ruleFields>

const rejectTenantOverlap = (values: RuleFields, context: z.RefinementCtx) => {
  if (values.was.some((id) => values.wasNot.includes(id))) {
    context.addIssue({
      code: 'custom',
      path: ['wasNot'],
      message: 'Арендатор не может быть одновременно в «был» и «не был»',
    })
  }
}

export const ruleFormSchema = ruleFields.superRefine(rejectTenantOverlap)

export const segmentFormSchema = ruleFields
  .extend({
    name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
    description: z.string().trim().max(1000, 'Не более 1000 символов'),
    templateCode: z.string().nullable(),
  })
  .superRefine(rejectTenantOverlap)

export type RuleFormValues = z.infer<typeof ruleFormSchema>
export type SegmentFormValues = z.infer<typeof segmentFormSchema>

export const emptyRuleValues: RuleFormValues = {
  periodDays: '90',
  was: [],
  wasNot: [],
  categoriesWas: [],
  categoriesWasNot: [],
  requireContact: 'none',
  minVisits: '',
  minAmount: '',
}

const nonEmpty = (ids: string[]) => (ids.length > 0 ? ids : undefined)

export function toRule(values: RuleFormValues): SegmentRule {
  return {
    periodDays: Number(values.periodDays),
    was: nonEmpty(values.was),
    wasNot: nonEmpty(values.wasNot),
    categoriesWas: nonEmpty(values.categoriesWas),
    categoriesWasNot: nonEmpty(values.categoriesWasNot),
    requireContact: values.requireContact === 'none' ? undefined : values.requireContact,
    minVisits: values.minVisits === '' ? undefined : Number(values.minVisits),
    minAmountMinor:
      values.minAmount === ''
        ? undefined
        : Math.round(Number(values.minAmount.replace(',', '.')) * MINOR_PER_MAJOR),
  }
}

export function toRuleFormValues(rule: SegmentRule): RuleFormValues {
  return {
    periodDays: String(rule.periodDays),
    was: rule.was ?? [],
    wasNot: rule.wasNot ?? [],
    categoriesWas: rule.categoriesWas ?? [],
    categoriesWasNot: rule.categoriesWasNot ?? [],
    requireContact: rule.requireContact ?? 'none',
    minVisits: rule.minVisits == null ? '' : String(rule.minVisits),
    minAmount: rule.minAmountMinor == null ? '' : String(rule.minAmountMinor / MINOR_PER_MAJOR),
  }
}
