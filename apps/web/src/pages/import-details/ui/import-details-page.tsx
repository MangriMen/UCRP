import { Link, useLocation, useParams } from 'react-router'
import { ImportStatusBadge, useImport } from '@/entities/import-batch'
import { formatDateTime } from '@/shared/lib/format'
import { Alert, AlertDescription } from '@/shared/ui/alert'
import { Button } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { BatchCounters } from './batch-counters'
import { RejectedTable } from './rejected-table'

export function ImportDetailsPage() {
  const { id = '' } = useParams()
  const location = useLocation()
  const { data: batch, error, isLoading } = useImport(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!batch) return <EmptyState title="Импорт не найден" />

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={batch.fileName ?? 'Синхронизация источника'}
        description={`${batch.dataSourceName ?? batch.dataSourceId} · ${formatDateTime(batch.createdAt)}`}
        actions={
          <>
            <ImportStatusBadge status={batch.status} />
            <Button variant="outline" render={<Link to="/import" />}>
              К списку
            </Button>
          </>
        }
      />
      {location.state?.isDuplicate && (
        <Alert>
          <AlertDescription>
            Такой файл уже загружали для этого источника, показан существующий импорт.
          </AlertDescription>
        </Alert>
      )}
      {batch.errorMessage && (
        <Alert variant="destructive">
          <AlertDescription>{batch.errorMessage}</AlertDescription>
        </Alert>
      )}
      <BatchCounters batch={batch} />
      {batch.rejectedRows > 0 && (
        <section className="flex flex-col gap-3">
          <h2 className="text-lg font-semibold">Отклонённые строки</h2>
          <RejectedTable batchId={batch.id} />
        </section>
      )}
    </div>
  )
}
