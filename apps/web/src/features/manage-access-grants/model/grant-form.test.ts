import { describe, expect, it } from 'vitest'
import {
  emptyGrantForm,
  fromLocalInput,
  toGrantPatch,
  toGrantRequest,
  validateGrantForm,
  type GrantFormValues,
} from './grant-form'
import type { SubjectValue } from './subject'

const now = new Date('2026-11-01T09:00:00Z')
const allow = { permission: 'tenant.aggregates.own', effect: 'allow' as const }

const tenant: SubjectValue = { kind: 'tenant', id: 'tenant-0001' }

const form = (overrides: Partial<GrantFormValues>): GrantFormValues => ({
  ...emptyGrantForm(now),
  items: [allow],
  reason: 'Пилот',
  ...overrides,
})

describe('grant form', () => {
  it('builds a tenant request with the mandatory tenant scope first', () => {
    const values = form({
      presetCode: 'tenant.self_aggregates',
      scopes: [{ type: 'period_days', value: ' 90 ' }],
    })
    expect(toGrantRequest(values, tenant, 'tenant-0001')).toEqual({
      subjectType: 'tenant',
      subjectId: 'tenant-0001',
      presetCode: 'tenant.self_aggregates',
      status: 'active',
      validFrom: fromLocalInput(values.validFrom),
      validTo: fromLocalInput(values.validTo),
      indefinite: false,
      reason: 'Пилот',
      items: [allow],
      scopes: [
        { type: 'tenant', value: 'tenant-0001' },
        { type: 'period_days', value: '90' },
      ],
    })
  })

  it('drops preset and scopes for a manager and sends an indefinite period without an end', () => {
    const manager: SubjectValue = { kind: 'manager', id: 'user-manager-2' }
    const values = form({
      presetCode: 'tenant.self_aggregates',
      scopes: [{ type: 'channel', value: 'sms' }],
      indefinite: true,
    })
    const request = toGrantRequest(values, manager, null)
    expect(request).toMatchObject({ subjectType: 'user', indefinite: true, scopes: [] })
    expect(request).not.toHaveProperty('presetCode')
    expect(request).not.toHaveProperty('validTo')
    expect(toGrantPatch(values, manager, null)).toMatchObject({
      presetCode: '',
      indefinite: true,
      scopes: [],
    })
  })

  it('targets the tenant role through the self scope', () => {
    const request = toGrantRequest(form({}), { kind: 'tenant_role', id: '' }, 'self')
    expect(request).toMatchObject({ subjectType: 'role', subjectId: 'tenant' })
    expect(request?.scopes[0]).toEqual({ type: 'tenant', value: 'self' })
  })

  it('reports missing parts of the grant', () => {
    const errors = validateGrantForm(
      form({ items: [], reason: ' ', validTo: '' }),
      { kind: 'tenant_user', id: 'user-x' },
      null,
    )
    expect(errors).toEqual({
      subject: 'Пользователь не привязан к арендатору',
      items: 'Выберите хотя бы одно право',
      validTo: 'Укажите окончание или выберите «Бессрочно»',
      reason: 'Укажите причину',
    })
  })

  it('rejects an end before the start and malformed numeric scopes', () => {
    const values = form({
      validFrom: '2026-11-10T10:00',
      validTo: '2026-11-01T10:00',
      scopes: [{ type: 'row_limit', value: 'много' }],
    })
    expect(validateGrantForm(values, tenant, 'tenant-0001')).toEqual({
      validTo: 'Окончание должно быть позже начала',
      scopes: 'Глубина периода, порог k и лимит строк задаются целым числом',
    })
  })
})
