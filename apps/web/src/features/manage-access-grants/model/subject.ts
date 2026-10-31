import type { AccessSubject, GrantableRole } from '@/entities/access'

export type SubjectKind = 'tenant' | 'tenant_user' | 'tenant_role' | 'manager'

export interface SubjectValue {
  kind: SubjectKind
  id: string
}

export interface DirectoryUser {
  id: string
  label: string
  role: string
  tenantId: string | null
}

export const TENANT_ROLE_ID = 'tenant'
export const SELF_TENANT_SCOPE = 'self'

export const subjectKindLabels: Record<SubjectKind, string> = {
  tenant: 'Арендатор',
  tenant_user: 'Пользователь арендатора',
  tenant_role: 'Все арендаторы (роль)',
  manager: 'Управляющий',
}

export const grantableRoleOf = (kind: SubjectKind): GrantableRole =>
  kind === 'manager' ? 'manager' : 'tenant'

export function toSubject({ kind, id }: SubjectValue): AccessSubject | null {
  if (kind === 'tenant_role') return { subjectType: 'role', subjectId: TENANT_ROLE_ID }
  if (!id) return null
  return { subjectType: kind === 'tenant' ? 'tenant' : 'user', subjectId: id }
}

export function fromSubject(subject: AccessSubject, users: DirectoryUser[]): SubjectValue {
  if (subject.subjectType === 'role') return { kind: 'tenant_role', id: '' }
  if (subject.subjectType === 'tenant') return { kind: 'tenant', id: subject.subjectId }
  const user = users.find((candidate) => candidate.id === subject.subjectId)
  return { kind: user?.role === 'manager' ? 'manager' : 'tenant_user', id: subject.subjectId }
}

export function tenantScopeOf(value: SubjectValue, users: DirectoryUser[]): string | null {
  switch (value.kind) {
    case 'tenant':
      return value.id || null
    case 'tenant_role':
      return SELF_TENANT_SCOPE
    case 'tenant_user':
      return users.find((user) => user.id === value.id)?.tenantId ?? null
    case 'manager':
      return null
  }
}
