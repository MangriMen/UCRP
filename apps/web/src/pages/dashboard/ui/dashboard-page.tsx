import { useState } from 'react'
import {
  MetricsPanel,
  PenetrationChart,
  granularityLabels,
  useDashboardMetrics,
  usePenetration,
  type Granularity,
} from '@/entities/dashboard'
import { defaultPeriod, isPeriodValid, periodQuery } from '@/shared/lib/period'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'
import { PeriodFilter } from '@/shared/ui/period-filter'
import { Skeleton } from '@/shared/ui/skeleton'

const granularityItems = Object.entries(granularityLabels).map(([value, label]) => ({
  value,
  label,
}))

export function DashboardPage() {
  const [period, setPeriod] = useState(() => defaultPeriod())
  const [granularity, setGranularity] = useState<Granularity>('month')
  const valid = isPeriodValid(period)
  const query = periodQuery(period)

  const metrics = useDashboardMetrics(query, valid)
  const penetration = usePenetration({ ...query, granularity }, valid)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Дашборд"
        description="Ключевые показатели потока данных, дедупликации и проникновения гостей"
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
          <CardTitle>Проникновение: гости по числу арендаторов</CardTitle>
        </CardHeader>
        <CardContent>
          {penetration.error ? (
            <ErrorBanner error={penetration.error} />
          ) : penetration.data ? (
            <PenetrationChart series={penetration.data.series} />
          ) : (
            valid && <Skeleton className="h-40 w-full" />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
