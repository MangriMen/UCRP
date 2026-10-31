import type { Me } from './types'

export type PermissionCode = string

export function hasPermission(user: Me | null, permission: PermissionCode): boolean {
  return user?.permissions.includes(permission) ?? false
}

export function hasAnyPermission(user: Me | null, permissions: PermissionCode[]): boolean {
  return permissions.some((permission) => hasPermission(user, permission))
}
