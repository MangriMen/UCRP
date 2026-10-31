import { describe, expect, it } from 'vitest'
import { auditSummary } from './audit-summary'
import type { AccessAuditEntry } from './types'

const entry = (action: string, details: unknown): AccessAuditEntry => ({
  id: 'audit-1',
  actorId: 'user-manager',
  actorRole: 'manager',
  action,
  entityId: 'grant-1',
  details,
  requestId: 'request-1',
  createdAt: '2026-11-01T10:00:00Z',
})

describe('auditSummary', () => {
  it('describes a status change with its reason', () => {
    const details = {
      before: { Status: 'active' },
      after: { Status: 'revoked', Reason: 'Договор' },
    }
    expect(auditSummary(entry('access.grant.revoke', details))).toBe(
      'Действует → Отозван · Причина: Договор',
    )
  })

  it('names the created status without a previous one', () => {
    expect(
      auditSummary(entry('access.grant.create', { before: null, after: { Status: 'draft' } })),
    ).toBe('Черновик')
  })

  it('describes a denied request', () => {
    const details = { method: 'GET', route: '/guests/{id}', code: 'forbidden_role' }
    expect(auditSummary(entry('access.denied', details))).toBe(
      'GET · /guests/{id} · forbidden_role',
    )
  })

  it('keeps plain text details', () => {
    expect(auditSummary(entry('access.grant.update', 'legacy'))).toBe('legacy')
  })
})
