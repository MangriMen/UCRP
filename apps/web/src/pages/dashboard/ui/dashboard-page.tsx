import { useState } from 'react'
import {
  MetricsPanel,
  CrossTenantChart,
  granularityLabels,
  useDashboardMetrics,
  useCrossTenantReport,
  type Granularity,
} from '@/entities/dashboard'
import { useAuth } from '@/entities/user'
import { useListParams } from '@/shared/lib/list-params'
import { defaultPeriod, isPeriodValid, periodQuery, type Period } from '@/shared/lib/period'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'
import { PeriodFilter } from '@/shared/ui/period-filter'
import { Skeleton } from '@/shared/ui/skeleton'
import { TenantBreakdown } from './tenant-breakdown'

const granularityItems = Object.entries(granularityLabels).map(([value, label]) => ({
  value,
  label,
}))

export function DashboardPage() {
  const { can } = useAuth()
  const { get, update } = useListParams()
  const [fallbackPeriod] = useState(() => defaultPeriod())
  const from = get('period_from')
  const to = get('period_to')
  const period: Period = from || to ? { from, to } : fallbackPeriod
  const setPeriod = (next: Period) => update({ period_from: next.from, period_to: next.to })
  const [granularity, setGranularity] = useState<Granularity>('month')
  const valid = isPeriodValid(period)
  const query = periodQuery(period)

  const metrics = useDashboardMetrics(query, valid)
  const crossTenant = useCrossTenantReport({ ...query, granularity }, valid)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Дашборд"
        description="Ключевые показатели потока данных, объединения профилей и перехода гостей между арендаторами"
      />
      <PeriodFilter
        value={period}
        onChange={setPeriod}
        error={valid ? undefined : 'Укажите корректный период'}
      >
        <FormField label="Детализация">
          {(props) => (
            <OptionSelect
              {...props}
              className="w-44"
              items={granularityItems}
              value={granularity}
              onValueChange={(value) => setGranularity(value as Granularity)}
            />
          )}
        </FormField>
      </PeriodFilter>
      {metrics.error ? (
        <ErrorBanner error={metrics.error} />
      ) : (
        <MetricsPanel metrics={metrics.data} isLoading={valid && metrics.isPending} />
      )}
      <Card>
        <CardHeader>
          <CardTitle>Гости по числу арендаторов</CardTitle>
        </CardHeader>
        <CardContent>
          {crossTenant.error ? (
            <ErrorBanner error={crossTenant.error} />
          ) : crossTenant.data ? (
            <CrossTenantChart series={crossTenant.data.series} />
          ) : (
            valid && <Skeleton className="h-40 w-full" />
          )}
        </CardContent>
      </Card>
      {valid && can('reference.read') && <TenantBreakdown period={period} />}
    </div>
  )
}
