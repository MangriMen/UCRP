import { useRouteError } from 'react-router'
import { Button } from '@/shared/ui/button'
import { ErrorBanner } from '@/shared/ui/error-banner'

export function RouteError() {
  const error = useRouteError()
  return (
    <div className="mx-auto flex max-w-lg flex-col gap-4 p-8">
      <ErrorBanner error={error} />
      <Button variant="outline" onClick={() => window.location.reload()}>
        Перезагрузить страницу
      </Button>
    </div>
  )
}
