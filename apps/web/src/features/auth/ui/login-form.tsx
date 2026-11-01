import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { useLogin } from '../api/use-login'

const schema = z.object({
  email: z.email('Введите корректный email'),
  password: z.string().min(1, 'Введите пароль'),
})

type LoginValues = z.infer<typeof schema>

interface LoginFormProps {
  onSuccess?: () => void
}

export function LoginForm({ onSuccess }: LoginFormProps) {
  const login = useLogin()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginValues>({ resolver: zodResolver(schema) })

  const submit = handleSubmit((values) => login.mutate(values, { onSuccess }))

  return (
    <form onSubmit={submit} className="flex flex-col gap-4" noValidate>
      {login.error && <ErrorBanner error={login.error} />}
      <FormField label="Email" error={errors.email?.message}>
        {(field) => (
          <Input type="email" autoComplete="username" {...field} {...register('email')} />
        )}
      </FormField>
      <FormField label="Пароль" error={errors.password?.message}>
        {(field) => (
          <Input
            type="password"
            autoComplete="current-password"
            {...field}
            {...register('password')}
          />
        )}
      </FormField>
      <Button type="submit" disabled={login.isPending}>
        Войти
      </Button>
    </form>
  )
}
