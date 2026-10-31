import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { hasErrorCode } from '@/shared/api/errors'
import { Alert, AlertDescription, AlertTitle } from '@/shared/ui/alert'
import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { useAcceptInvitation } from '../api/use-accept-invitation'
import { acceptInvitationSchema, type AcceptInvitationValues } from '../model/schema'

interface AcceptInvitationFormProps {
  token: string
  onSuccess: () => void
}

export function AcceptInvitationForm({ token, onSuccess }: AcceptInvitationFormProps) {
  const accept = useAcceptInvitation()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<AcceptInvitationValues>({ resolver: zodResolver(acceptInvitationSchema) })

  const submit = handleSubmit(({ password }) => accept.mutate({ token, password }, { onSuccess }))

  if (hasErrorCode(accept.error, 'unauthorized')) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Ссылка недействительна</AlertTitle>
        <AlertDescription>
          Срок действия ссылки истёк, она уже использована или неверна. Попросите управляющего
          отправить новое приглашение.
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-4" noValidate>
      {accept.error && <ErrorBanner error={accept.error} />}
      <FormField label="Новый пароль" error={errors.password?.message} hint="От 8 до 72 символов">
        {(field) => (
          <Input type="password" autoComplete="new-password" {...field} {...register('password')} />
        )}
      </FormField>
      <FormField label="Повторите пароль" error={errors.confirmation?.message}>
        {(field) => (
          <Input
            type="password"
            autoComplete="new-password"
            {...field}
            {...register('confirmation')}
          />
        )}
      </FormField>
      <Button type="submit" disabled={accept.isPending}>
        Установить пароль
      </Button>
    </form>
  )
}
