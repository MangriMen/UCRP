import { formatPercent } from '@/shared/lib/format'
import { Alert, AlertDescription } from '@/shared/ui/alert'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Skeleton } from '@/shared/ui/skeleton'
import { useCampaignResults } from '../api/use-campaigns'

function Metric({ title, value }: { title: string; value: string | number }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-normal text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent className="text-2xl font-semibold">{value}</CardContent>
    </Card>
  )
}

export function CampaignResults({ campaignId }: { campaignId: string }) {
  const { data, error, isLoading } = useCampaignResults(campaignId)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!data) return <EmptyState title="Результатов пока нет" />

  return (
    <div className="flex flex-col gap-4">
      {data.isSignificant != null && (
        <Alert variant={data.isSignificant ? 'default' : 'destructive'}>
          <AlertDescription>
            {data.isSignificant
              ? 'Прирост статистически значим'
              : 'Прирост статистически не значим'}
            {data.pValue != null && ` (p = ${data.pValue.toFixed(3)})`}
          </AlertDescription>
        </Alert>
      )}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Metric title="Получатели" value={data.recipients} />
        <Metric title="Конверсия получателей" value={formatPercent(data.conversionRate)} />
        <Metric title="Контрольная группа" value={data.controlGroup} />
        <Metric
          title="Конверсия контроля"
          value={
            data.controlConversionRate != null ? formatPercent(data.controlConversionRate) : '—'
          }
        />
        <Metric title="Конверсии" value={data.conversions} />
        <Metric
          title="Прирост конверсии"
          value={data.incrementalRate != null ? formatPercent(data.incrementalRate) : '—'}
        />
        <Metric title="Дополнительные визиты" value={data.incrementalVisits ?? '—'} />
      </div>
    </div>
  )
}
