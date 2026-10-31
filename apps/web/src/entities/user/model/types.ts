import type { Schema } from '@/shared/api/types'

export type Me = Schema<'MeResponse'>
export type Role = Schema<'UserRole'>
export type User = Schema<'User'>
export type UserCreateRequest = Schema<'UserCreateRequest'>
export type UserUpdateRequest = Schema<'UserUpdateRequest'>
