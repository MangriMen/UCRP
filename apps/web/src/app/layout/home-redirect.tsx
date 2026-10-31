import { Navigate } from 'react-router'
import { useAuth } from '@/entities/user'
import { navigationFor } from '../navigation'

export function HomeRedirect() {
  const { user } = useAuth()
  const first = user ? navigationFor(user)[0]?.items[0] : undefined
  if (!first) return null
  return <Navigate to={first.to} replace />
}
