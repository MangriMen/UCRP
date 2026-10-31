import { useState } from 'react'
import {
  formatAggregateCell,
  useTenantAggregates,
  type TenantAggregatePoint,
  type Granularity,
} from '@/entities/tenant-aggregates'
import { useAuth } from '@/entities/user'
import { defaultPeriod, isPeriodValid, periodQuery } from '@/shared/lib/period'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'
import { PeriodFilter } from '@/shared/ui/period-filter'
import { Skeleton } from '@/shared/ui/skeleton'
import { SmallCellNotice } from './small-cell-notice'

const DEFAULT_THRESHOLD = 5

const granularityItems = [
  { value: 'day', label: 'По дням' },
  { value: 'week', label: 'По неделям' },
  { value: 'month', label: 'По месяцам' },
]

export function TenantAggregatesPage() {
  const { user } = useAuth()
  const tenantId = user?.tenantId ?? null
  const [period, setPeriod] = useState(() => defaultPeriod())
  const [granularity, setGranularity] = useState<Granularity>('month')
  const valid = isPeriodValid(period)

  const aggregates = useTenantAggregates(valid ? tenantId : null, {
    ...periodQuery(period),
    granularity,
  })

  if (!tenantId) {
    return <EmptyState title="Арендатор не привязан к вашей учётной записи" />
  }

  const data = aggregates.data
  const threshold = data?.suppressionThreshold ?? DEFAULT_THRESHOLD
  const cell = (value: number | null | undefined) => formatAggregateCell(value, threshold)

  const columns: Column<TenantAggregatePoint>[] = [
    { id: 'period', header: 'Период', cell: (p) => p.period },
    { id: 'newGuests', header: 'Новые гости', cell: (p) => cell(p.newGuests) },
    { id: 'repeatVisits', header: 'Повторные визиты', cell: (p) => cell(p.repeatVisits) },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={data ? `Мой арендатор: ${data.tenantName}` : 'Мой арендатор'}
        description="Агрегированная статистика без данных об отдельных гостях"
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
      {data?.suppressed && <SmallCellNotice threshold={threshold} />}
      {data && (
        <div className="grid gap-4 sm:grid-cols-2">
          <Card size="sm">
            <CardHeader>
              <CardTitle>Новые гости через комплекс</CardTitle>
            </CardHeader>
            <CardContent className="text-3xl font-semibold" data-testid="new-guests">
              {cell(data.newGuestsViaComplex)}
            </CardContent>
          </Card>
          <Card size="sm">
            <CardHeader>
              <CardTitle>Повторные визиты</CardTitle>
            </CardHeader>
            <CardContent className="text-3xl font-semibold" data-testid="repeat-visits">
              {cell(data.repeatVisits)}
            </CardContent>
          </Card>
        </div>
      )}
      {valid && aggregates.isPending && <Skeleton className="h-24 w-full" />}
      {aggregates.error ? (
        <ErrorBanner error={aggregates.error} />
      ) : (
        data && (
          <DataTable
            columns={columns}
            rows={data.series}
            getRowId={(p) => p.period}
            emptyTitle="Нет данных за период"
          />
        )
      )}
    </div>
  )
}
