import { useSearchParams } from 'react-router'
import {
  formatAggregateCell,
  formatAggregateMoney,
  isVisitAggregate,
  useTenantAggregates,
  type TenantAggregateDay,
  type TenantVisitAggregate,
} from '@/entities/tenant-aggregates'
import { useAuth } from '@/entities/user'
import { daysAgo, isPeriodValid, type Period } from '@/shared/lib/period'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PeriodFilter } from '@/shared/ui/period-filter'
import { Skeleton } from '@/shared/ui/skeleton'
import { StatCard, StatGrid } from '@/shared/ui/stat-card'
import { ActivityChart, type ActivityMetric, type ActivityPoint } from './activity-chart'
import { SmallCellNotice } from './small-cell-notice'

const NEW_GUESTS_METRIC = 'metric.new_guests_via_mall.read'
const REPEAT_VISITS_METRIC = 'metric.repeat_visits.read'
const FIRST_PURCHASES_METRIC = 'metric.first_purchase_new_tenant.read'
const REVENUE_METRIC = 'metric.revenue_total.read'
const AVG_CHECK_METRIC = 'metric.avg_check.read'
const DEFAULT_PERIOD_DAYS = 29
const MAX_PERIOD_DAYS = 366
const DAY_MS = 86_400_000

type AggregateValue = (data: TenantVisitAggregate) => number | null | undefined

interface OverviewCard {
  id: string
  label: string
  hint?: string
  permission?: string
  format: (data: TenantVisitAggregate) => string
}

const count =
  (value: AggregateValue) =>
  (data: TenantVisitAggregate): string =>
    formatAggregateCell(value(data), data.suppressionThreshold)

const money =
  (value: AggregateValue) =>
  (data: TenantVisitAggregate): string =>
    formatAggregateMoney(value(data), data.suppressionThreshold)

const overviewCards: OverviewCard[] = [
  { id: 'visits-count', label: 'Визиты', format: count((data) => data.visitsCount) },
  { id: 'guests-count', label: 'Гостей', format: count((data) => data.guestsCount) },
  {
    id: 'new-guests-count',
    label: 'Новые гости',
    permission: NEW_GUESTS_METRIC,
    format: count((data) => data.newGuestsCount),
  },
  {
    id: 'repeat-visits-count',
    label: 'Повторные визиты',
    hint: 'Визиты гостей, которые уже бывали у арендатора',
    permission: REPEAT_VISITS_METRIC,
    format: count((data) => data.repeatVisitsCount),
  },
  {
    id: 'first-purchases-count',
    label: 'Первые покупки',
    hint: 'Гости, чья первая покупка у арендатора пришлась на период',
    permission: FIRST_PURCHASES_METRIC,
    format: count((data) => data.firstPurchasesCount),
  },
  {
    id: 'revenue-total',
    label: 'Выручка',
    permission: REVENUE_METRIC,
    format: money((data) => data.revenueTotalMinor),
  },
  {
    id: 'avg-check',
    label: 'Средний чек',
    permission: AVG_CHECK_METRIC,
    format: money((data) => data.avgCheckMinor),
  },
]

const revenuePoints = (series: TenantAggregateDay[]): ActivityPoint[] =>
  series.map((day) => ({
    date: day.date,
    revenue:
      day.revenueTotalMinor === null || day.revenueTotalMinor === undefined
        ? null
        : day.revenueTotalMinor / 100,
  }))

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

  const cards = overviewCards.filter((card) => !card.permission || can(card.permission))
  const activityMetrics: ActivityMetric[] = [
    'visitsCount',
    ...(can(NEW_GUESTS_METRIC) ? (['newGuestsCount'] as const) : []),
    ...(can(REPEAT_VISITS_METRIC) ? (['repeatVisitsCount'] as const) : []),
  ]
  const showRevenueChart = can(REVENUE_METRIC)

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
            {cards.map((card) => (
              <StatCard
                key={card.id}
                label={card.label}
                hint={card.hint}
                value={<span data-testid={card.id}>{card.format(data)}</span>}
              />
            ))}
          </StatGrid>
          {data.series.length > 0 && (
            <ActivityChart
              title="Динамика по дням"
              points={data.series}
              metrics={activityMetrics}
            />
          )}
          {data.series.length > 0 && showRevenueChart && (
            <ActivityChart
              title="Выручка по дням"
              points={revenuePoints(data.series)}
              metrics={['revenue']}
            />
          )}
        </>
      )}
      {isValid && data && !isVisitAggregate(data) && (
        <EmptyState title="Показатели пока не открыты управляющим" />
      )}
    </div>
  )
}
