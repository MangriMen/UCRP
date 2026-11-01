import { describe, expect, it } from 'vitest'
import {
  emptyGroupValues,
  emptyLeafValues,
  MAX_GROUP_DEPTH,
  ruleFormSchema,
  segmentFormSchema,
  toRule,
  toRuleFormValues,
  zoneKey,
} from './schema'

const group = (overrides: Partial<ReturnType<typeof emptyGroupValues>> = {}) => ({
  ...emptyGroupValues(),
  ...overrides,
})

describe('segment rule form', () => {
  it('maps form values to a group node without empty conditions', () => {
    const values = group({
      conditions: [{ ...emptyLeafValues, visitedTenantIds: ['t1'], minAmount: '50,5' }],
    })
    expect(toRule(values)).toEqual({
      operator: 'and',
      conditions: [{ periodDays: 90, visitedTenantIds: ['t1'], minAmountMinor: 5050 }],
    })
  })

  it('loads a legacy flat rule as a root AND group with one leaf', () => {
    const rule = {
      periodDays: 30,
      excludedTenantIds: ['t2'],
      requireContact: 'email' as const,
      minVisits: 3,
    }
    const values = toRuleFormValues(rule)
    expect(values.operator).toBe('and')
    expect(values.conditions).toHaveLength(1)
    expect(values.groups).toEqual([])
    expect(toRule(values)).toEqual({ operator: 'and', conditions: [rule] })
  })

  it('maps selected zones to purchasedZones and omits an empty selection', () => {
    const withZones = group({
      conditions: [{ ...emptyLeafValues, purchasedZones: [zoneKey({ mallId: 'm1', floor: 2 })] }],
    })
    expect(toRule(withZones)).toEqual({
      operator: 'and',
      conditions: [{ periodDays: 90, purchasedZones: [{ mallId: 'm1', floor: 2 }] }],
    })
    expect(toRule(group())).toEqual({ operator: 'and', conditions: [{ periodDays: 90 }] })
  })

  it('keeps negative floors and mall ids with separators when mapping zones', () => {
    const zones = [
      { mallId: 'mall|north', floor: -1 },
      { mallId: 'm2', floor: 3 },
    ]
    const rule = {
      operator: 'and' as const,
      conditions: [{ periodDays: 30, purchasedZones: zones }],
    }
    expect(toRule(toRuleFormValues(rule))).toEqual(rule)
  })

  it('round-trips a nested group', () => {
    const rule = {
      operator: 'and' as const,
      conditions: [
        { periodDays: 90, requireContact: 'phone' as const },
        {
          operator: 'or' as const,
          conditions: [
            { periodDays: 30, minVisits: 2 },
            { periodDays: 60, visitedTenantIds: ['t1'] },
          ],
        },
      ],
    }
    expect(toRule(toRuleFormValues(rule))).toEqual(rule)
  })

  it('accepts a nested group within the depth limit', () => {
    expect(MAX_GROUP_DEPTH).toBe(1)
    const targeted = { ...emptyLeafValues, visitedTenantIds: ['t1'] }
    const values = group({
      conditions: [targeted],
      groups: [group({ operator: 'or', conditions: [targeted] })],
    })
    expect(ruleFormSchema.safeParse(values).success).toBe(true)
  })

  it('rejects nesting deeper than the limit', () => {
    const values = group({ groups: [group({ groups: [group()] })] })
    expect(ruleFormSchema.safeParse(values).success).toBe(false)
  })

  it('rejects an empty group', () => {
    expect(ruleFormSchema.safeParse(group({ conditions: [] })).success).toBe(false)
    expect(ruleFormSchema.safeParse(group({ groups: [group({ conditions: [] })] })).success).toBe(
      false,
    )
  })

  it('rejects a condition without a tenant, category or purchase zone', () => {
    const result = ruleFormSchema.safeParse(group())
    expect(result.success).toBe(false)
    expect(result.error?.issues[0]).toMatchObject({
      path: ['conditions', 0, 'visitedTenantIds'],
      message: 'Выберите арендатора, категорию или этаж покупки',
    })
  })

  it('rejects a period outside the contract range', () => {
    const leaf = (periodDays: string) => group({ conditions: [{ ...emptyLeafValues, periodDays }] })
    expect(ruleFormSchema.safeParse(leaf('731')).success).toBe(false)
    expect(ruleFormSchema.safeParse(leaf('0')).success).toBe(false)
  })

  it('rejects a tenant listed in both visitedTenantIds and excludedTenantIds', () => {
    const result = ruleFormSchema.safeParse(
      group({
        conditions: [{ ...emptyLeafValues, visitedTenantIds: ['t1'], excludedTenantIds: ['t1'] }],
      }),
    )
    expect(result.error?.issues[0]?.path).toEqual(['conditions', 0, 'excludedTenantIds'])
  })

  it('requires a name', () => {
    const result = segmentFormSchema.safeParse({
      ...emptyGroupValues(),
      name: ' ',
      description: '',
      templateCode: null,
    })
    expect(result.error?.issues[0]?.path).toEqual(['name'])
  })
})
