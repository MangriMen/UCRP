import { z } from 'zod'
import {
  isRuleGroup,
  type SegmentRule,
  type SegmentRuleNode,
  type ZoneRef,
} from '@/entities/segment'

export const MAX_PERIOD_DAYS = 730

export const contactOptions = ['none', 'phone', 'email'] as const

const MINOR_PER_MAJOR = 100

const isIntegerInRange = (value: string, min: number, max: number) => {
  if (!/^\d+$/.test(value)) return false
  const number = Number(value)
  return number >= min && number <= max
}

const leafFields = z.object({
  periodDays: z
    .string()
    .trim()
    .refine((value) => isIntegerInRange(value, 1, MAX_PERIOD_DAYS), {
      message: `Целое число от 1 до ${MAX_PERIOD_DAYS}`,
    }),
  visitedTenantIds: z.array(z.string()),
  excludedTenantIds: z.array(z.string()),
  visitedCategoryIds: z.array(z.string()),
  excludedCategoryIds: z.array(z.string()),
  purchasedZones: z.array(z.string()),
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

type LeafValues = z.infer<typeof leafFields>

const rejectTenantOverlap = (values: LeafValues, context: z.RefinementCtx) => {
  if (values.visitedTenantIds.some((id) => values.excludedTenantIds.includes(id))) {
    context.addIssue({
      code: 'custom',
      path: ['excludedTenantIds'],
      message: 'Арендатор не может быть одновременно в «был» и «не был»',
    })
  }
}

const requireTarget = (values: LeafValues, context: z.RefinementCtx) => {
  const hasTarget =
    values.visitedTenantIds.length > 0 ||
    values.visitedCategoryIds.length > 0 ||
    values.purchasedZones.length > 0
  if (!hasTarget) {
    context.addIssue({
      code: 'custom',
      path: ['visitedTenantIds'],
      message: 'Выберите арендатора, категорию или этаж покупки',
    })
  }
}

export const leafSchema = leafFields.superRefine((values, context) => {
  rejectTenantOverlap(values, context)
  requireTarget(values, context)
})

export const MAX_GROUP_DEPTH = 1

export const operatorOptions = ['and', 'or'] as const

export type Operator = (typeof operatorOptions)[number]

export interface GroupValues {
  operator: Operator
  conditions: LeafValues[]
  groups: GroupValues[]
}

const buildGroupSchema = (depth: number): z.ZodType<GroupValues, GroupValues> =>
  z.object({
    operator: z.enum(operatorOptions),
    conditions: z.array(leafSchema).min(1, 'Добавьте хотя бы одно условие'),
    groups: depth > 0 ? z.array(buildGroupSchema(depth - 1)) : z.array(z.never()).max(0),
  })

export const ruleFormSchema = buildGroupSchema(MAX_GROUP_DEPTH)

export const segmentFormSchema = z.object({
  name: z.string().trim().min(1, 'Введите название').max(200, 'Не более 200 символов'),
  description: z.string().trim().max(1000, 'Не более 1000 символов'),
  templateCode: z.string().nullable(),
  operator: z.enum(operatorOptions),
  conditions: z.array(leafSchema).min(1, 'Добавьте хотя бы одно условие'),
  groups: z.array(buildGroupSchema(MAX_GROUP_DEPTH - 1)),
})

export type LeafFormValues = LeafValues
export type RuleFormValues = GroupValues
export type SegmentFormValues = z.infer<typeof segmentFormSchema>

export const emptyLeafValues: LeafFormValues = {
  periodDays: '90',
  visitedTenantIds: [],
  excludedTenantIds: [],
  visitedCategoryIds: [],
  excludedCategoryIds: [],
  purchasedZones: [],
  requireContact: 'none',
  minVisits: '',
  minAmount: '',
}

export const emptyGroupValues = (): GroupValues => ({
  operator: 'and',
  conditions: [{ ...emptyLeafValues }],
  groups: [],
})

const nonEmpty = <T>(items: T[]) => (items.length > 0 ? items : undefined)

const ZONE_KEY_SEPARATOR = '|'

export const zoneKey = (zone: ZoneRef) => `${zone.mallId}${ZONE_KEY_SEPARATOR}${zone.floor}`

const parseZoneKey = (key: string): ZoneRef => {
  const separatorIndex = key.lastIndexOf(ZONE_KEY_SEPARATOR)
  return {
    mallId: key.slice(0, separatorIndex),
    floor: Number(key.slice(separatorIndex + 1)),
  }
}

function toLeaf(values: LeafFormValues): SegmentRule {
  return {
    periodDays: Number(values.periodDays),
    visitedTenantIds: nonEmpty(values.visitedTenantIds),
    excludedTenantIds: nonEmpty(values.excludedTenantIds),
    visitedCategoryIds: nonEmpty(values.visitedCategoryIds),
    excludedCategoryIds: nonEmpty(values.excludedCategoryIds),
    purchasedZones: nonEmpty(values.purchasedZones.map(parseZoneKey)),
    requireContact: values.requireContact === 'none' ? undefined : values.requireContact,
    minVisits: values.minVisits === '' ? undefined : Number(values.minVisits),
    minAmountMinor:
      values.minAmount === ''
        ? undefined
        : Math.round(Number(values.minAmount.replace(',', '.')) * MINOR_PER_MAJOR),
  }
}

export function toRule(values: GroupValues): SegmentRuleNode {
  return {
    operator: values.operator,
    conditions: [...values.conditions.map(toLeaf), ...values.groups.map(toRule)],
  }
}

function toLeafValues(rule: SegmentRule): LeafFormValues {
  return {
    periodDays: String(rule.periodDays),
    visitedTenantIds: rule.visitedTenantIds ?? [],
    excludedTenantIds: rule.excludedTenantIds ?? [],
    visitedCategoryIds: rule.visitedCategoryIds ?? [],
    excludedCategoryIds: rule.excludedCategoryIds ?? [],
    purchasedZones: (rule.purchasedZones ?? []).map(zoneKey),
    requireContact: rule.requireContact ?? 'none',
    minVisits: rule.minVisits == null ? '' : String(rule.minVisits),
    minAmount: rule.minAmountMinor == null ? '' : String(rule.minAmountMinor / MINOR_PER_MAJOR),
  }
}

const isLeaf = (node: SegmentRuleNode): node is SegmentRule => !isRuleGroup(node)

function toGroupValues(node: SegmentRuleNode, depth: number): GroupValues {
  if (!isRuleGroup(node)) {
    return { operator: 'and', conditions: [toLeafValues(node)], groups: [] }
  }
  const leaves = node.conditions.filter(isLeaf)
  const groups = depth > 0 ? node.conditions.filter(isRuleGroup) : []
  return {
    operator: node.operator,
    conditions: leaves.map(toLeafValues),
    groups: groups.map((group) => toGroupValues(group, depth - 1)),
  }
}

export const toRuleFormValues = (node: SegmentRuleNode): GroupValues =>
  toGroupValues(node, MAX_GROUP_DEPTH)
