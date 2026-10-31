import { http, HttpResponse } from 'msw'
import { mockTenants } from '@/entities/tenant/mocks'
import { mockUsers } from '@/entities/user/mocks'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import { transitionSources, type GrantTransition } from '../model/lifecycle'
import type {
  AccessAuditEntry,
  AccessGrant,
  AccessGrantPatch,
  AccessGrantRequest,
  AccessSubject,
  GrantStatus,
} from '../model/types'
import {
  mockAccessAudit,
  mockAccessGrants,
  mockAccessPermissions,
  mockAccessPresets,
} from './fixtures'

const grants: AccessGrant[] = structuredClone(mockAccessGrants)
const auditLog: AccessAuditEntry[] = structuredClone(mockAccessAudit)
let policyVersion = 3

const SESSION_KEY = 'ucrp-mock-session'
const sessionUserId = () => localStorage.getItem(SESSION_KEY) ?? 'user-manager'

const transitionTargets: Record<GrantTransition, GrantStatus> = {
  activate: 'active',
  suspend: 'suspended',
  revoke: 'revoked',
}

const subjectUser = (id: string) => mockUsers.find((user) => user.userId === id)

function subjectProblem({ subjectType, subjectId }: AccessSubject) {
  if (subjectType === 'tenant' && !mockTenants.some((tenant) => tenant.id === subjectId)) {
    return apiError(404, 'not_found', 'Не найдено')
  }
  if (subjectType === 'role' && subjectId !== 'tenant')
    return apiError(403, 'forbidden_role', 'Нет прав')
  if (subjectType === 'user') {
    const user = subjectUser(subjectId)
    if (!user) return apiError(404, 'not_found', 'Не найдено')
    if (user.role !== 'tenant' && user.role !== 'manager')
      return apiError(403, 'forbidden_role', 'Нет прав')
  }
  return null
}

function requestProblem(body: AccessGrantRequest) {
  const problem = subjectProblem(body)
  if (problem) return problem
  const isManager = subjectUser(body.subjectId)?.role === 'manager'
  if (isManager && body.subjectId === sessionUserId())
    return apiError(403, 'forbidden_role', 'Нет прав')
  const hasEnd = Boolean(body.validTo)
  if (
    !body.reason.trim() ||
    body.items.length === 0 ||
    hasEnd === Boolean(body.indefinite) ||
    (isManager && (body.scopes.length > 0 || body.presetCode))
  ) {
    return apiError(400, 'validation_error', 'Некорректные данные')
  }
  return null
}

const record = (action: string, entityId: string, details: unknown) => {
  auditLog.unshift({
    id: mockId('access-audit-new'),
    actorId: sessionUserId(),
    actorRole: subjectUser(sessionUserId())?.role ?? 'manager',
    action,
    entityId,
    details,
    requestId: 'mock-request',
    createdAt: new Date().toISOString(),
  })
}

const appliesTo = (grant: AccessGrant, { subjectType, subjectId }: AccessSubject) => {
  if (grant.subjectType === subjectType && grant.subjectId === subjectId) return true
  if (subjectType === 'role') return false
  const tenantId = subjectType === 'tenant' ? subjectId : subjectUser(subjectId)?.tenantId
  if (!tenantId) return false
  return (
    (grant.subjectType === 'tenant' && grant.subjectId === tenantId) ||
    (grant.subjectType === 'role' && grant.subjectId === 'tenant')
  )
}

function effective(subject: AccessSubject, extra: AccessGrant[] = []) {
  const user = subject.subjectType === 'user' ? subjectUser(subject.subjectId) : undefined
  const base = user?.role === 'manager' ? user.permissions : []
  const active = [...grants, ...extra].filter(
    (grant) => grant.status === 'active' && appliesTo(grant, subject),
  )
  const allowed = new Set(base)
  const denied = new Set<string>()
  for (const grant of active) {
    for (const item of grant.items) {
      if (item.effect === 'allow') allowed.add(item.permission)
      else denied.add(item.permission)
    }
  }
  return {
    subjectType: subject.subjectType,
    subjectId: subject.subjectId,
    policyVersion,
    permissions: [...allowed].filter((code) => !denied.has(code)),
    userGrantsIncluded: subject.subjectType === 'user',
  }
}

const findGrant = (id: unknown) => grants.find((grant) => grant.id === id)

export const accessHandlers = [
  http.get(`${API}/access/permissions`, () => HttpResponse.json({ items: mockAccessPermissions })),

  http.get(`${API}/access/presets`, () => HttpResponse.json({ items: mockAccessPresets })),

  http.get(`${API}/access/grants`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const subjectType = params.get('subject_type')
    const subjectId = params.get('subject_id')
    const status = params.get('status')
    const filtered = grants.filter(
      (grant) =>
        (!subjectType || grant.subjectType === subjectType) &&
        (!subjectId || grant.subjectId === subjectId) &&
        (!status || grant.status === status),
    )
    return HttpResponse.json(paged(request, filtered))
  }),

  http.get(`${API}/access/grants/:id`, ({ params }) => {
    const grant = findGrant(params.id)
    return grant ? HttpResponse.json(grant) : apiError(404, 'not_found', 'Не найдено')
  }),

  http.post(`${API}/access/grants`, async ({ request }) => {
    const body = (await request.json()) as AccessGrantRequest
    const problem = requestProblem(body)
    if (problem) return problem
    const now = new Date().toISOString()
    policyVersion += 1
    const grant: AccessGrant = {
      id: mockId('grant-new'),
      subjectType: body.subjectType,
      subjectId: body.subjectId,
      presetCode: body.presetCode || undefined,
      status: body.status,
      validFrom: body.validFrom,
      validTo: body.validTo ?? null,
      createdBy: sessionUserId(),
      reason: body.reason,
      policyVersion,
      createdAt: now,
      updatedAt: now,
      items: body.items,
      scopes: body.scopes,
    }
    grants.unshift(grant)
    record('access.grant.create', grant.id, {
      before: null,
      after: { Status: grant.status, Reason: grant.reason },
    })
    return HttpResponse.json(grant, { status: 201 })
  }),

  http.patch(`${API}/access/grants/:id`, async ({ params, request }) => {
    const grant = findGrant(params.id)
    if (!grant) return apiError(404, 'not_found', 'Не найдено')
    if (grant.status === 'revoked' || grant.status === 'expired') {
      return apiError(409, 'conflict', 'Конфликт состояния')
    }
    const body = (await request.json()) as AccessGrantPatch
    if (body.reason !== undefined && !body.reason.trim()) {
      return apiError(400, 'validation_error', 'Некорректные данные')
    }
    const before = { Status: grant.status, Reason: grant.reason }
    Object.assign(grant, {
      presetCode: body.presetCode === undefined ? grant.presetCode : body.presetCode || undefined,
      validFrom: body.validFrom ?? grant.validFrom,
      validTo: body.indefinite ? null : (body.validTo ?? grant.validTo),
      reason: body.reason ?? grant.reason,
      items: body.items ?? grant.items,
      scopes: body.scopes ?? grant.scopes,
      updatedAt: new Date().toISOString(),
    })
    record('access.grant.update', grant.id, {
      before,
      after: { Status: grant.status, Reason: grant.reason },
    })
    return HttpResponse.json(grant)
  }),

  http.post(`${API}/access/grants/:id/:transition`, async ({ params, request }) => {
    const grant = findGrant(params.id)
    const transition = params.transition as GrantTransition
    if (!grant || !(transition in transitionSources))
      return apiError(404, 'not_found', 'Не найдено')
    const { reason } = (await request.json()) as { reason: string }
    if (!reason?.trim()) return apiError(400, 'validation_error', 'Некорректные данные')
    if (!transitionSources[transition].includes(grant.status)) {
      return apiError(409, 'conflict', 'Конфликт состояния')
    }
    const before = { Status: grant.status, Reason: grant.reason }
    grant.status = transitionTargets[transition]
    grant.reason = reason
    grant.updatedAt = new Date().toISOString()
    policyVersion += 1
    record(`access.grant.${transition}`, grant.id, {
      before,
      after: { Status: grant.status, Reason: reason },
    })
    return HttpResponse.json(grant)
  }),

  http.get(`${API}/access/effective`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const subject = {
      subjectType: params.get('subject_type') as AccessSubject['subjectType'],
      subjectId: params.get('subject_id') ?? '',
    }
    return subjectProblem(subject) ?? HttpResponse.json(effective(subject))
  }),

  http.post(`${API}/access/simulate`, async ({ request }) => {
    const body = (await request.json()) as AccessGrantRequest
    const problem = requestProblem(body)
    if (problem) return problem
    const simulated: AccessGrant = {
      ...body,
      id: 'simulation',
      status: 'active',
      validTo: body.validTo ?? null,
      createdBy: sessionUserId(),
      policyVersion,
      createdAt: body.validFrom,
      updatedAt: body.validFrom,
    }
    record('access.simulate', body.subjectId, {
      subjectType: body.subjectType,
      subjectId: body.subjectId,
    })
    return HttpResponse.json(effective(body, [simulated]))
  }),

  http.get(`${API}/access/audit`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const grantId = params.get('grant_id')
    const actor = params.get('actor')
    const from = params.get('from')
    const to = params.get('to')
    const filtered = auditLog.filter(
      (entry) =>
        (!grantId || (entry.entityId === grantId && entry.action.startsWith('access.grant.'))) &&
        (!actor || entry.actorId === actor) &&
        (!from || entry.createdAt >= from) &&
        (!to || entry.createdAt <= to),
    )
    return HttpResponse.json(paged(request, filtered))
  }),
]
