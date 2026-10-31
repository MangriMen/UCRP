import { ApiError } from '@/shared/api/client'
import { Alert, AlertDescription, AlertTitle } from '@/shared/ui/alert'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { mergeConflictMessages } from '../model/labels'

export function MergeErrorAlert({ error }: { error: unknown }) {
  if (!(error instanceof ApiError) || error.status !== 409) return <ErrorBanner error={error} />
  return (
    <Alert variant="destructive">
      <AlertTitle>Конфликт</AlertTitle>
      <AlertDescription>
        {mergeConflictMessages[error.body.code] ?? error.body.message}
      </AlertDescription>
    </Alert>
  )
}
