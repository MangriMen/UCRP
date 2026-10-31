import type { ReactNode } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router'
import { EmptyState } from '@/shared/ui/empty-state'
import { Skeleton } from '@/shared/ui/skeleton'
import { useAuth } from '../api/use-auth'
import type { Role } from '../model/types'

interface RoleGuardProps {
  roles?: Role[]
  children?: ReactNode
}

export function RoleGuard({ roles, children }: RoleGuardProps) {
  const { user, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) return <Skeleton className="m-4 h-8 w-64" />
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />
  if (roles && !roles.includes(user.role)) {
    return (
      <EmptyState title="Недостаточно прав" description="Этот раздел недоступен для вашей роли." />
    )
  }
  return children ?? <Outlet />
}
