import { ApiError, type ApiErrorCode } from '@/shared/api/client'
import { Alert, AlertDescription, AlertTitle } from './alert'

const titleByCode: Partial<Record<ApiErrorCode, string>> = {
  forbidden_role: 'Недостаточно прав',
  forbidden_tenant: 'Недостаточно прав',
  forbidden_field: 'Недостаточно прав',
  not_found: 'Не найдено',
  conflict: 'Конфликт данных',
  duplicate: 'Такая запись уже есть',
  validation_error: 'Проверьте введённые данные',
  timeout: 'Сервер не ответил вовремя',
  rate_limited: 'Слишком много запросов',
}

interface ErrorBannerProps {
  error: unknown
}

export function ErrorBanner({ error }: ErrorBannerProps) {
  if (!(error instanceof ApiError)) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Что-то пошло не так</AlertTitle>
        <AlertDescription>Не удалось выполнить запрос. Попробуйте ещё раз.</AlertDescription>
      </Alert>
    )
  }

  const { code, message, details, request_id } = error.body
  return (
    <Alert variant="destructive">
      <AlertTitle>{titleByCode[code] ?? 'Ошибка запроса'}</AlertTitle>
      <AlertDescription>
        <p>{message}</p>
        {details && details.length > 0 && (
          <ul className="mt-1 list-disc pl-4">
            {details.map((d) => (
              <li key={`${d.field}:${d.reason}`}>
                {d.field}: {d.reason}
              </li>
            ))}
          </ul>
        )}
        {request_id && <p className="mt-1 text-xs opacity-70">ID запроса: {request_id}</p>}
      </AlertDescription>
    </Alert>
  )
}
