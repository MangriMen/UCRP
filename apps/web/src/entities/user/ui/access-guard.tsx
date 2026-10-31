import type { ReactNode } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router'
import { EmptyState } from '@/shared/ui/empty-state'
import { Skeleton } from '@/shared/ui/skeleton'
import { useAuth } from '../api/use-auth'
import type { PermissionCode } from '../model/permissions'

interface AccessGuardProps {
  permission?: PermissionCode
  tenantArea?: boolean
  children?: ReactNode
}

export function AccessGuard({ permission, tenantArea = false, children }: AccessGuardProps) {
  const { user, isLoading, can } = useAuth()
  const location = useLocation()

  if (isLoading) return <Skeleton className="m-4 h-8 w-64" />
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />
  const allowed = !permission || (can(permission) && (tenantArea || user.role !== 'tenant'))
  if (!allowed) {
    return <EmptyState title="Недостаточно прав" description="Этот раздел вам недоступен." />
  }
  return children ?? <Outlet />
}
