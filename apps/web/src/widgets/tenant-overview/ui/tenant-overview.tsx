import { useSearchParams } from 'react-router'
import {
  formatAggregateCell,
  isVisitAggregate,
  useTenantAggregates,
} from '@/entities/tenant-aggregates'
import { useAuth } from '@/entities/user'
import { daysAgo, isPeriodValid, type Period } from '@/shared/lib/period'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PeriodFilter } from '@/shared/ui/period-filter'
import { Skeleton } from '@/shared/ui/skeleton'
import { StatCard, StatGrid } from '@/shared/ui/stat-card'
import { ActivityChart } from './activity-chart'
import { SmallCellNotice } from './small-cell-notice'

const NEW_GUESTS_METRIC = 'metric.new_guests_via_mall.read'
const DEFAULT_PERIOD_DAYS = 29
const MAX_PERIOD_DAYS = 366
const DAY_MS = 86_400_000

const periodLength = ({ from, to }: Period) => (Date.parse(to) - Date.parse(from)) / DAY_MS + 1

const periodError = (period: Period) => {
  if (period.from === '' || period.to === '') return 'Укажите начало и конец периода'
  if (period.from > period.to) return 'Начало периода позже конца'
  if (periodLength(period) > MAX_PERIOD_DAYS) return 'Период не может быть длиннее 366 дней'
  return undefined
}

interface TenantOverviewProps {
  tenantId: string
}

export function TenantOverview({ tenantId }: TenantOverviewProps) {
  const { can } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const period: Period = {
    from: searchParams.get('from') ?? daysAgo(DEFAULT_PERIOD_DAYS),
    to: searchParams.get('to') ?? daysAgo(0),
  }
  const error = periodError(period)
  const isValid = error === undefined && isPeriodValid(period)
  const aggregates = useTenantAggregates(isValid ? tenantId : null, isValid ? period : undefined)
  const showNewGuests = can(NEW_GUESTS_METRIC)

  const changePeriod = ({ from, to }: Period) =>
    setSearchParams(
      (current) => {
        const next = new URLSearchParams(current)
        if (from) next.set('from', from)
        else next.delete('from')
        if (to) next.set('to', to)
        else next.delete('to')
        return next
      },
      { replace: true },
    )

  const data = aggregates.data

  return (
    <div className="flex flex-col gap-4">
      <PeriodFilter value={period} onChange={changePeriod} error={error} />
      {isValid && aggregates.isPending && <Skeleton className="h-24 w-full" />}
      {isValid && aggregates.error && <ErrorBanner error={aggregates.error} />}
      {isValid && data && isVisitAggregate(data) && (
        <>
          {data.suppressed && <SmallCellNotice threshold={data.suppressionThreshold} />}
          <StatGrid>
            <StatCard
              label="Визиты"
              value={
                <span data-testid="visits-count">
                  {formatAggregateCell(data.visitsCount, data.suppressionThreshold)}
                </span>
              }
            />
            <StatCard
              label="Гостей"
              value={
                <span data-testid="guests-count">
                  {formatAggregateCell(data.guestsCount, data.suppressionThreshold)}
                </span>
              }
            />
            {showNewGuests && (
              <StatCard
                label="Новые гости"
                value={
                  <span data-testid="new-guests-count">
                    {formatAggregateCell(data.newGuestsCount, data.suppressionThreshold)}
                  </span>
                }
              />
            )}
          </StatGrid>
          {data.series.length > 0 && (
            <ActivityChart series={data.series} showNewGuests={showNewGuests} />
          )}
        </>
      )}
      {isValid && data && !isVisitAggregate(data) && (
        <EmptyState title="Показатели пока не открыты управляющим" />
      )}
    </div>
  )
}
