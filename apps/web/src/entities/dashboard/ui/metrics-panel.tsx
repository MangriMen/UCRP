import { formatNumber, formatPercent } from '@/shared/lib/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { Skeleton } from '@/shared/ui/skeleton'
import type { DashboardMetrics } from '../model/types'

const formatSeconds = (seconds: number) => {
  const minutes = Math.floor(seconds / 60)
  const rest = Math.round(seconds % 60)
  return minutes > 0 ? `${minutes} мин ${rest} с` : `${rest} с`
}

interface Metric {
  id: string
  title: string
  value: string
  target: string
  sample?: string
}

function sampleText(label: string, value: number | undefined) {
  return value === undefined ? undefined : `${label}: ${formatNumber(value)}`
}

function buildMetrics({
  matchRate,
  disputedMergeRate,
  conversionRate,
  conversionLift,
  segmentToExportMedianSeconds,
  segmentToExportP95Seconds,
  sampleSizes,
}: DashboardMetrics): Metric[] {
  const ratio = (value: number | null | undefined) => (value == null ? '—' : formatPercent(value))
  const duration = (value: number | null | undefined) =>
    value == null ? '—' : formatSeconds(value)
  return [
    {
      id: 'matchRate',
      title: 'Доля распознанных событий',
      value: ratio(matchRate),
      target: 'Цель: не менее 70%',
      sample: sampleText('событий', sampleSizes?.events),
    },
    {
      id: 'disputedMergeRate',
      title: 'Доля оспоренных объединений',
      value: ratio(disputedMergeRate),
      target: 'Цель: менее 1%',
      sample: sampleText('кандидатов на объединение', sampleSizes?.mergeCandidates),
    },
    {
      id: 'segmentToExport',
      title: 'Сегмент → выгрузка (медиана)',
      value: duration(segmentToExportMedianSeconds),
      target: `Цель: до 5 мин, p95: ${duration(segmentToExportP95Seconds)}`,
      sample: sampleText('выгрузок', sampleSizes?.exports),
    },
    {
      id: 'conversionRate',
      title: 'Конверсия кампаний',
      value: ratio(conversionRate),
      target: 'Цель: не менее 3%',
      sample: sampleText('контактов', sampleSizes?.contacts),
    },
    {
      id: 'conversionLift',
      title: 'Прирост к контрольной группе',
      value: ratio(conversionLift),
      target: 'Цель: больше 0',
      sample: sampleText('гостей', sampleSizes?.guests),
    },
  ]
}

interface MetricsPanelProps {
  metrics: DashboardMetrics | undefined
  isLoading?: boolean
}

export function MetricsPanel({ metrics, isLoading }: MetricsPanelProps) {
  if (isLoading || !metrics) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {Array.from({ length: 5 }, (_, index) => (
          <Skeleton key={index} className="h-32 w-full" />
        ))}
      </div>
    )
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {buildMetrics(metrics).map((metric) => (
        <Card key={metric.id} size="sm">
          <CardHeader>
            <CardTitle>{metric.title}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-1">
            <p className="text-3xl font-semibold" data-testid={`metric-${metric.id}`}>
              {metric.value}
            </p>
            <p className="text-xs text-muted-foreground">{metric.target}</p>
            {metric.sample && <p className="text-xs text-muted-foreground">{metric.sample}</p>}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
