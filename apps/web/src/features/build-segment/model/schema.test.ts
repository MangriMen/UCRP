import { describe, expect, it } from 'vitest'
import {
  emptyRuleValues,
  ruleFormSchema,
  segmentFormSchema,
  toRule,
  toRuleFormValues,
} from './schema'

describe('segment rule form', () => {
  it('maps form values to a rule without empty conditions', () => {
    expect(toRule({ ...emptyRuleValues, was: ['t1'], minAmount: '50,5' })).toEqual({
      periodDays: 90,
      was: ['t1'],
      minAmountMinor: 5050,
    })
  })

  it('round-trips a rule through the form', () => {
    const rule = { periodDays: 30, wasNot: ['t2'], requireContact: 'email' as const, minVisits: 3 }
    expect(toRule(toRuleFormValues(rule))).toEqual(rule)
  })

  it('rejects a period outside the contract range', () => {
    expect(ruleFormSchema.safeParse({ ...emptyRuleValues, periodDays: '731' }).success).toBe(false)
    expect(ruleFormSchema.safeParse({ ...emptyRuleValues, periodDays: '0' }).success).toBe(false)
  })

  it('rejects a tenant listed in both was and wasNot', () => {
    const result = ruleFormSchema.safeParse({ ...emptyRuleValues, was: ['t1'], wasNot: ['t1'] })
    expect(result.error?.issues[0]?.path).toEqual(['wasNot'])
  })

  it('requires a name', () => {
    const result = segmentFormSchema.safeParse({
      ...emptyRuleValues,
      name: ' ',
      description: '',
      templateCode: null,
    })
    expect(result.error?.issues[0]?.path).toEqual(['name'])
  })
})
