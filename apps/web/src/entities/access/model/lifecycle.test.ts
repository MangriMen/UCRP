import { describe, expect, it } from 'vitest'
import { availableTransitions, isGrantEditable } from './lifecycle'
import type { AccessGrant } from './types'

const now = new Date('2026-11-01T12:00:00Z')

const grant = (overrides: Partial<AccessGrant>): AccessGrant => ({
  id: 'grant-1',
  subjectType: 'tenant',
  subjectId: 'tenant-0001',
  status: 'active',
  validFrom: '2026-10-01T00:00:00Z',
  validTo: null,
  createdBy: 'user-manager',
  reason: 'Пилот',
  policyVersion: 1,
  createdAt: '2026-10-01T00:00:00Z',
  updatedAt: '2026-10-01T00:00:00Z',
  items: [],
  scopes: [],
  ...overrides,
})

describe('grant lifecycle', () => {
  it('offers the transitions the backend accepts for each status', () => {
    expect(availableTransitions(grant({ status: 'draft' }), now)).toEqual(['activate', 'revoke'])
    expect(availableTransitions(grant({ status: 'active' }), now)).toEqual(['suspend', 'revoke'])
    expect(availableTransitions(grant({ status: 'suspended' }), now)).toEqual([
      'activate',
      'revoke',
    ])
    expect(availableTransitions(grant({ status: 'revoked' }), now)).toEqual([])
    expect(availableTransitions(grant({ status: 'expired' }), now)).toEqual([])
  })

  it('freezes staff role grants but not the tenant role grant', () => {
    const staff = grant({ subjectType: 'role', subjectId: 'marketer' })
    expect(availableTransitions(staff, now)).toEqual([])
    expect(isGrantEditable(staff, now)).toBe(false)

    const tenantRole = grant({ subjectType: 'role', subjectId: 'tenant' })
    expect(availableTransitions(tenantRole, now)).toEqual(['suspend', 'revoke'])
    expect(isGrantEditable(tenantRole, now)).toBe(true)
  })

  it('freezes a grant whose end has passed', () => {
    const ended = grant({ validTo: '2026-11-01T11:00:00Z' })
    expect(availableTransitions(ended, now)).toEqual([])
    expect(isGrantEditable(ended, now)).toBe(false)
    expect(isGrantEditable(grant({ status: 'suspended' }), now)).toBe(true)
  })
})
