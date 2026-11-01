import { useState } from 'react'
import { useSearchParams } from 'react-router'
import { AcceptInvitationForm } from '@/features/accept-invitation'
import { Alert, AlertDescription, AlertTitle } from '@/shared/ui/alert'
import { ButtonLink } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'

function LoginLink() {
  return (
    <ButtonLink to="/login" variant="outline" className="w-full">
      Перейти ко входу
    </ButtonLink>
  )
}

function AcceptInvitationContent({ token }: { token: string }) {
  const [accepted, setAccepted] = useState(false)

  if (!token) {
    return (
      <div className="flex flex-col gap-4">
        <Alert variant="destructive">
          <AlertTitle>Ссылка неполная</AlertTitle>
          <AlertDescription>
            В ссылке нет кода приглашения. Откройте ссылку из письма целиком.
          </AlertDescription>
        </Alert>
        <LoginLink />
      </div>
    )
  }

  if (accepted) {
    return (
      <div className="flex flex-col gap-4">
        <Alert>
          <AlertTitle>Пароль установлен</AlertTitle>
          <AlertDescription>Войдите с email из письма и новым паролем.</AlertDescription>
        </Alert>
        <LoginLink />
      </div>
    )
  }

  return <AcceptInvitationForm token={token} onSuccess={() => setAccepted(true)} />
}

export function AcceptInvitationPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Установка пароля</CardTitle>
        </CardHeader>
        <CardContent>
          <AcceptInvitationContent token={token} />
        </CardContent>
      </Card>
    </div>
  )
}
