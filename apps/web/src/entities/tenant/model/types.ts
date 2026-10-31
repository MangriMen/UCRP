import type { Schema } from '@/shared/api/types'

export type Tenant = Schema<'Tenant'>
export type TenantListItem = Schema<'TenantListItem'>
export type TenantCreateRequest = Schema<'TenantCreateRequest'>
export type TenantUpdateRequest = Schema<'TenantUpdateRequest'>
export type TenantListSort =
  | 'name'
  | '-name'
  | 'floor'
  | '-floor'
  | 'priority'
  | '-priority'
  | 'createdAt'
  | '-createdAt'
  | 'guestsCount'
  | '-guestsCount'
  | 'visitsCount'
  | '-visitsCount'
  | 'newGuestsCount'
  | '-newGuestsCount'
