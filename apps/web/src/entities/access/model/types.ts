import type { Schema } from '@/shared/api/types'

export type AccessGrant = Schema<'Grant'>
export type AccessGrantRequest = Schema<'GrantRequest'>
export type AccessGrantPatch = Schema<'GrantPatch'>
export type AccessGrantItem = Schema<'GrantItem'>
export type AccessGrantScope = Schema<'GrantScope'>
export type AccessPermission = Schema<'Permission'>
export type AccessPreset = Schema<'Preset'>
export type SubjectPermissions = Schema<'SubjectPermissions'>
export type AccessAuditEntry = Schema<'AuditEntry'>

export type GrantStatus = AccessGrant['status']
export type GrantSubjectType = AccessGrant['subjectType']
export type GrantScopeType = AccessGrantScope['type']
export type GrantEffect = AccessGrantItem['effect']
export type GrantableRole = AccessPermission['grantableTo'][number]

export interface AccessSubject {
  subjectType: GrantSubjectType
  subjectId: string
}
