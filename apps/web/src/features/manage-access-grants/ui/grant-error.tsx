import { ApiError } from '@/shared/api/client'
import { ErrorBanner } from '@/shared/ui/error-banner'

const hintByStatus: Record<number, string> = {
  400: 'Право не входит в потолок роли субъекта, нет обязательной области арендатора или срок задан неверно.',
  403: 'Выдать или снять эти права от вашего имени нельзя: нужен неограниченный доступ к каждому праву на весь срок гранта. Себе выдавать гранты тоже нельзя.',
  404: 'Субъект или грант не найден.',
  409: 'Грант уже в другом статусе, истёк или политика доступа изменилась. Обновите страницу.',
}

export function GrantError({ error }: { error: unknown }) {
  const hint = error instanceof ApiError ? hintByStatus[error.status] : undefined
  return (
    <div className="flex flex-col gap-1">
      <ErrorBanner error={error} />
      {hint && <p className="text-sm text-muted-foreground">{hint}</p>}
    </div>
  )
}
