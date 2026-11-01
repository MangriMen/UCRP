import { Navigate, useLocation } from 'react-router'
import { returnPathFor, useAuth } from '@/entities/user'
import { LoginForm } from '@/features/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'

export function LoginPage() {
  const { user } = useAuth()
  const from = (useLocation().state as { from?: string } | null)?.from ?? '/'

  if (user) return <Navigate to={returnPathFor(user, from)} replace />

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Вход в UCRP</CardTitle>
        </CardHeader>
        <CardContent>
          <LoginForm />
        </CardContent>
      </Card>
    </div>
  )
}
