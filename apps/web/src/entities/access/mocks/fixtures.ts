import type {
  AccessAuditEntry,
  AccessGrant,
  AccessPermission,
  AccessPreset,
  GrantableRole,
} from '../model/types'

const permission = (
  code: string,
  title: string,
  grantableTo: GrantableRole[],
  kind: AccessPermission['kind'] = 'action',
): AccessPermission => ({
  code,
  domain: code.split('.')[0],
  kind,
  title,
  description: '',
  isSensitive: code.startsWith('guest.') || code.startsWith('field.'),
  minK: kind === 'metric' ? 5 : null,
  grantableTo,
})

export const mockAccessPermissions: AccessPermission[] = [
  permission('tenant.aggregates.own', 'Агрегаты только своего арендатора', ['tenant']),
  permission('reference.read', 'Справочники', ['tenant']),
  permission('metric.visits_count.read', 'Число визитов', ['tenant'], 'metric'),
  permission(
    'metric.new_guests_via_mall.read',
    'Новые гости, пришедшие через комплекс',
    ['tenant'],
    'metric',
  ),
  permission('metric.repeat_visits.read', 'Повторные визиты', ['tenant'], 'metric'),
  permission('metric.revenue_total.read', 'Суммарная выручка', [], 'metric'),
  permission('guest.profile.read', 'Профиль гостя', ['tenant', 'manager']),
  permission('guest.profile.search', 'Поиск гостей', ['manager']),
  permission('guest.event.read', 'Лента событий', ['tenant', 'manager']),
  permission('guest.identifier.read', 'Идентификаторы гостя', ['tenant', 'manager']),
  permission('field.guest.phone.read', 'Телефон гостя', ['tenant', 'manager'], 'field'),
  permission(
    'field.guest.identifiers.read',
    'Идентификаторы целиком',
    ['tenant', 'manager'],
    'field',
  ),
  permission('field.event.amount.read', 'Сумма события', ['tenant', 'manager'], 'field'),
  permission('access.grant.write', 'Выдача и отзыв грантов', []),
  permission('audit.read', 'Журнал аудита', []),
]

const allow = (codes: string[]) =>
  codes.map((code) => ({ permission: code, effect: 'allow' as const }))

export const mockAccessPresets: AccessPreset[] = [
  { code: 'tenant.none', items: [], available: false },
  {
    code: 'tenant.self_aggregates',
    items: allow([
      'tenant.aggregates.own',
      'metric.new_guests_via_mall.read',
      'metric.repeat_visits.read',
      'metric.visits_count.read',
      'reference.read',
    ]),
    available: true,
  },
  {
    code: 'tenant.self_aggregates_extended',
    items: allow(['tenant.aggregates.own', 'metric.segment_size.read']),
    available: false,
  },
  {
    code: 'tenant.own_events',
    items: allow(['guest.event.read', 'field.event.amount.read']),
    available: true,
  },
  {
    code: 'tenant.full_profile_read',
    items: allow([
      'guest.profile.read',
      'guest.identifier.read',
      'field.guest.identifiers.read',
      'field.guest.phone.read',
    ]),
    available: true,
  },
]

const grant = (overrides: Partial<AccessGrant> & Pick<AccessGrant, 'id'>): AccessGrant => ({
  subjectType: 'tenant',
  subjectId: 'tenant-0001',
  status: 'active',
  validFrom: '2026-09-01T00:00:00Z',
  validTo: null,
  createdBy: 'user-manager',
  reason: 'Пилот',
  policyVersion: 3,
  createdAt: '2026-09-01T08:00:00Z',
  updatedAt: '2026-09-01T08:00:00Z',
  items: [],
  scopes: [],
  ...overrides,
})

export const mockAccessGrants: AccessGrant[] = [
  grant({
    id: 'grant-0006',
    subjectId: 'tenant-0003',
    status: 'revoked',
    reason: 'Договор расторгнут',
    items: allow(['metric.visits_count.read']),
    scopes: [{ type: 'tenant', value: 'tenant-0003' }],
    createdAt: '2026-10-20T08:00:00Z',
    updatedAt: '2026-10-25T08:00:00Z',
  }),
  grant({
    id: 'grant-0005',
    subjectType: 'user',
    subjectId: 'user-manager',
    createdBy: '',
    reason: 'Initial personal data access migration',
    items: allow(['guest.profile.read', 'guest.profile.search', 'guest.event.read']),
    createdAt: '2026-10-15T08:00:00Z',
    updatedAt: '2026-10-15T08:00:00Z',
  }),
  grant({
    id: 'grant-0004',
    subjectType: 'role',
    subjectId: 'tenant',
    reason: 'Справочники для всех арендаторов',
    items: allow(['reference.read']),
    scopes: [{ type: 'tenant', value: 'self' }],
    createdAt: '2026-10-10T08:00:00Z',
    updatedAt: '2026-10-10T08:00:00Z',
  }),
  grant({
    id: 'grant-0003',
    subjectType: 'user',
    subjectId: 'user-tenant',
    status: 'suspended',
    validTo: '2027-06-30T00:00:00Z',
    reason: 'Проверка жалобы гостя',
    items: allow(['guest.profile.read']),
    scopes: [{ type: 'tenant', value: 'tenant-0001' }],
    createdAt: '2026-10-05T08:00:00Z',
    updatedAt: '2026-10-07T08:00:00Z',
  }),
  grant({
    id: 'grant-0002',
    subjectId: 'tenant-0002',
    status: 'draft',
    presetCode: 'tenant.own_events',
    validTo: '2027-01-31T00:00:00Z',
    reason: 'События своих гостей',
    items: allow(['guest.event.read', 'field.event.amount.read']),
    scopes: [
      { type: 'tenant', value: 'tenant-0002' },
      { type: 'period_days', value: '90' },
    ],
    createdAt: '2026-09-20T08:00:00Z',
    updatedAt: '2026-09-20T08:00:00Z',
  }),
  grant({
    id: 'grant-0001',
    presetCode: 'tenant.self_aggregates',
    reason: 'Пилот: агрегаты для арендатора',
    items: allow([
      'tenant.aggregates.own',
      'metric.new_guests_via_mall.read',
      'metric.repeat_visits.read',
      'metric.visits_count.read',
      'reference.read',
    ]),
    scopes: [{ type: 'tenant', value: 'tenant-0001' }],
  }),
]

const audit = (overrides: Partial<AccessAuditEntry> & Pick<AccessAuditEntry, 'id'>) => ({
  actorId: 'user-manager',
  actorRole: 'manager',
  action: 'access.grant.create',
  entityId: 'grant-0001',
  details: null,
  requestId: 'mock-request',
  createdAt: '2026-09-01T08:00:00Z',
  ...overrides,
})

export const mockAccessAudit: AccessAuditEntry[] = [
  audit({
    id: 'access-audit-0004',
    action: 'access.denied',
    actorId: 'user-tenant',
    actorRole: 'tenant',
    entityId: 'user-tenant',
    details: { method: 'GET', route: '/guests/{id}', code: 'forbidden_permission' },
    createdAt: '2026-10-28T09:30:00Z',
  }),
  audit({
    id: 'access-audit-0003',
    action: 'access.grant.revoke',
    entityId: 'grant-0006',
    details: {
      before: { Status: 'active' },
      after: { Status: 'revoked', Reason: 'Договор расторгнут' },
    },
    createdAt: '2026-10-25T08:00:00Z',
  }),
  audit({
    id: 'access-audit-0002',
    action: 'access.simulate',
    entityId: 'tenant-0002',
    details: { subjectType: 'tenant', subjectId: 'tenant-0002' },
    createdAt: '2026-09-20T07:55:00Z',
  }),
  audit({
    id: 'access-audit-0001',
    details: {
      before: null,
      after: { Status: 'active', Reason: 'Пилот: агрегаты для арендатора' },
    },
  }),
]
