import { hasAnyPermission, hasPermission, type PermissionCode } from '../model/permissions'
import { useMe } from './use-me'

export function useAuth() {
  const { data, isPending } = useMe()
  const user = data ?? null
  return {
    user,
    role: user?.role ?? null,
    isLoading: isPending,
    can: (permission: PermissionCode) => hasPermission(user, permission),
    canAny: (permissions: PermissionCode[]) => hasAnyPermission(user, permissions),
  }
}
