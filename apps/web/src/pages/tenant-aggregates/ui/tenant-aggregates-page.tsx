import {
  formatAggregateCell,
  isVisitAggregate,
  useTenantAggregates,
} from '@/entities/tenant-aggregates'
import { useAuth } from '@/entities/user'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { SmallCellNotice } from './small-cell-notice'

const DEFAULT_THRESHOLD = 5

export function TenantAggregatesPage() {
  const { user } = useAuth()
  const tenantId = user?.tenantId ?? null
  const aggregates = useTenantAggregates(tenantId)

  if (!tenantId) {
    return <EmptyState title="Арендатор не привязан к вашей учётной записи" />
  }

  const data = aggregates.data

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Мой арендатор"
        description="Агрегированная статистика без данных об отдельных гостях"
      />
      {aggregates.isPending && <Skeleton className="h-24 w-full" />}
      {aggregates.error && <ErrorBanner error={aggregates.error} />}
      {data && isVisitAggregate(data) && (
        <>
          {data.suppressed && <SmallCellNotice threshold={DEFAULT_THRESHOLD} />}
          <Card size="sm">
            <CardHeader>
              <CardTitle>Визиты</CardTitle>
            </CardHeader>
            <CardContent className="text-3xl font-semibold" data-testid="visits-count">
              {formatAggregateCell(data.visitsCount, DEFAULT_THRESHOLD)}
            </CardContent>
          </Card>
        </>
      )}
      {data && !isVisitAggregate(data) && (
        <EmptyState title="Показатели пока не открыты управляющим" />
      )}
    </div>
  )
}
