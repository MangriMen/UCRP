import { describe, expect, it } from 'vitest'
import {
  emptyGroupValues,
  emptyLeafValues,
  MAX_GROUP_DEPTH,
  ruleFormSchema,
  segmentFormSchema,
  toRule,
  toRuleFormValues,
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
    const values = group({ groups: [group({ operator: 'or' })] })
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
