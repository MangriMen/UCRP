import { Navigate } from 'react-router'
import { useAuth } from '@/entities/user'
import { navigationFor } from '../navigation'

export function HomeRedirect() {
  const { role } = useAuth()
  const first = role ? navigationFor(role)[0]?.items[0] : undefined
  if (!first) return null
  return <Navigate to={first.to} replace />
}
