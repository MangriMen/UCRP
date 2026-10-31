import { Navigate } from 'react-router'
import { useAuth } from '@/entities/user'
import { EmptyState } from '@/shared/ui/empty-state'
import { navigationFor } from '../navigation'

export function HomeRedirect() {
  const { user } = useAuth()
  const first = user ? navigationFor(user)[0]?.items[0] : undefined
  if (!first) {
    return (
      <EmptyState
        title="Нет доступных разделов"
        description="Для вашей учётной записи не выданы права на разделы интерфейса"
      />
    )
  }
  return <Navigate to={first.to} replace />
}
