import type { ReactNode } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router'
import { ButtonLink } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'
import { Skeleton } from '@/shared/ui/skeleton'
import { useAuth } from '../api/use-auth'
import { hasAllPermissions, type PermissionCode } from '../model/permissions'

interface AccessGuardProps {
  permission?: PermissionCode | PermissionCode[]
  tenantArea?: boolean
  children?: ReactNode
}

export function AccessGuard({ permission, tenantArea = false, children }: AccessGuardProps) {
  const { user, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) return <Skeleton className="m-4 h-8 w-64" />
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />
  const allowed =
    !permission ||
    (hasAllPermissions(user, [permission].flat()) && (tenantArea || user.role !== 'tenant'))
  if (!allowed && tenantArea && user.role === 'tenant') return <TenantAccessPending />
  if (!allowed) {
    return (
      <EmptyState
        title="Недостаточно прав"
        description="Этот раздел вам недоступен."
        action={
          <ButtonLink to="/" variant="outline" size="sm">
            На стартовую страницу
          </ButtonLink>
        }
      />
    )
  }
  return children ?? <Outlet />
}

export function TenantAccessPending() {
  return (
    <EmptyState
      title="Доступ ещё не выдан"
      description="Обратитесь к управляющему торгового центра: он выдаёт доступ к данным вашего арендатора."
    />
  )
}
