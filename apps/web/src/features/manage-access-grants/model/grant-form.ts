import type {
  AccessGrant,
  AccessGrantItem,
  AccessGrantPatch,
  AccessGrantRequest,
  AccessGrantScope,
  GrantScopeType,
} from '@/entities/access'
import { toSubject, type SubjectValue } from './subject'

export type ExtraScopeType = Exclude<GrantScopeType, 'tenant'>

export interface ScopeDraft {
  type: ExtraScopeType
  value: string
}

export interface GrantFormValues {
  presetCode: string
  items: AccessGrantItem[]
  scopes: ScopeDraft[]
  validFrom: string
  validTo: string
  indefinite: boolean
  status: 'draft' | 'active'
  reason: string
}

export type GrantFormErrors = Partial<
  Record<'subject' | 'items' | 'scopes' | 'validFrom' | 'validTo' | 'reason', string>
>

export const EXTRA_SCOPE_TYPES: ExtraScopeType[] = [
  'category',
  'period_days',
  'channel',
  'k_threshold',
  'row_limit',
  'field_mask',
  'metric',
]

const NUMERIC_SCOPES: ExtraScopeType[] = ['period_days', 'k_threshold', 'row_limit']
const DEFAULT_DURATION_DAYS = 30
const DAY_MS = 86_400_000

const pad = (value: number) => String(value).padStart(2, '0')

export function toLocalInput(iso: string) {
  const date = new Date(iso)
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export const fromLocalInput = (value: string) => new Date(value).toISOString()

export function emptyGrantForm(now = new Date()): GrantFormValues {
  return {
    presetCode: '',
    items: [],
    scopes: [],
    validFrom: toLocalInput(now.toISOString()),
    validTo: toLocalInput(new Date(now.getTime() + DEFAULT_DURATION_DAYS * DAY_MS).toISOString()),
    indefinite: false,
    status: 'active',
    reason: '',
  }
}

export function grantFormFromGrant(grant: AccessGrant): GrantFormValues {
  const scopes = grant.scopes.filter(
    (scope): scope is AccessGrantScope & { type: ExtraScopeType } => scope.type !== 'tenant',
  )
  return {
    presetCode: grant.presetCode ?? '',
    items: grant.items,
    scopes,
    validFrom: toLocalInput(grant.validFrom),
    validTo: grant.validTo ? toLocalInput(grant.validTo) : '',
    indefinite: !grant.validTo,
    status: grant.status === 'draft' ? 'draft' : 'active',
    reason: grant.reason,
  }
}

function scopeProblem(scopes: ScopeDraft[]) {
  const seen = new Set<string>()
  for (const scope of scopes) {
    const value = scope.value.trim()
    if (!value) return 'Заполните значения ограничений'
    if (NUMERIC_SCOPES.includes(scope.type) && !/^\d+$/.test(value)) {
      return 'Глубина периода, порог k и лимит строк задаются целым числом'
    }
    const key = `${scope.type}:${value}`
    if (seen.has(key)) return 'Ограничения повторяются'
    seen.add(key)
  }
  return undefined
}

export function validateGrantForm(
  values: GrantFormValues,
  subject: SubjectValue,
  tenantScope: string | null,
): GrantFormErrors {
  const errors: GrantFormErrors = {}
  if (!toSubject(subject)) errors.subject = 'Выберите субъект'
  else if (subject.kind !== 'manager' && !tenantScope) {
    errors.subject = 'Пользователь не привязан к арендатору'
  }
  if (values.items.length === 0) errors.items = 'Выберите хотя бы одно право'
  const scopes = scopeProblem(values.scopes)
  if (scopes) errors.scopes = scopes
  if (!values.validFrom) errors.validFrom = 'Укажите начало действия'
  if (!values.indefinite) {
    if (!values.validTo) errors.validTo = 'Укажите окончание или выберите «Бессрочно»'
    else if (values.validFrom && values.validTo <= values.validFrom) {
      errors.validTo = 'Окончание должно быть позже начала'
    }
  }
  if (!values.reason.trim()) errors.reason = 'Укажите причину'
  return errors
}

function requestScopes(
  values: GrantFormValues,
  subject: SubjectValue,
  tenantScope: string | null,
): AccessGrantScope[] {
  if (subject.kind === 'manager' || !tenantScope) return []
  return [
    { type: 'tenant', value: tenantScope },
    ...values.scopes.map((scope) => ({ type: scope.type, value: scope.value.trim() })),
  ]
}

const period = (values: GrantFormValues) =>
  values.indefinite
    ? { validFrom: fromLocalInput(values.validFrom), indefinite: true }
    : {
        validFrom: fromLocalInput(values.validFrom),
        validTo: fromLocalInput(values.validTo),
        indefinite: false,
      }

const presetOf = (values: GrantFormValues, subject: SubjectValue) =>
  subject.kind === 'manager' ? '' : values.presetCode

export function toGrantRequest(
  values: GrantFormValues,
  subject: SubjectValue,
  tenantScope: string | null,
): AccessGrantRequest | null {
  const target = toSubject(subject)
  if (!target) return null
  const presetCode = presetOf(values, subject)
  return {
    ...target,
    ...(presetCode ? { presetCode } : {}),
    ...period(values),
    status: values.status,
    reason: values.reason.trim(),
    items: values.items,
    scopes: requestScopes(values, subject, tenantScope),
  }
}

export function toGrantPatch(
  values: GrantFormValues,
  subject: SubjectValue,
  tenantScope: string | null,
): AccessGrantPatch {
  return {
    presetCode: presetOf(values, subject),
    ...period(values),
    reason: values.reason.trim(),
    items: values.items,
    scopes: requestScopes(values, subject, tenantScope),
  }
}
