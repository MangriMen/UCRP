import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import { formatPercent } from '@/shared/lib/format'
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/shared/ui/chart'
import { EmptyState } from '@/shared/ui/empty-state'
import type { PenetrationPoint } from '../model/types'

const chartConfig = {
  one: { label: '1 арендатор', color: 'var(--chart-3)' },
  two: { label: '2 арендатора', color: 'var(--chart-2)' },
  threePlus: { label: '3 и более', color: 'var(--chart-1)' },
} satisfies ChartConfig

const segmentKeys = ['one', 'two', 'threePlus'] as const

interface PenetrationChartProps {
  series: PenetrationPoint[]
}

function describePoint(point: PenetrationPoint) {
  const rate =
    point.penetrationRate == null
      ? ''
      : `, доля двух и более: ${formatPercent(point.penetrationRate)}`
  return `${point.period}: ${point.one} / ${point.two} / ${point.threePlus}${rate}`
}

export function PenetrationChart({ series }: PenetrationChartProps) {
  if (series.length === 0) return <EmptyState title="Нет данных за период" />

  const ariaLabel = `Распределение гостей по числу арендаторов. ${series.map(describePoint).join('; ')}`

  return (
    <figure className="flex flex-col gap-3">
      <ChartContainer
        config={chartConfig}
        role="img"
        aria-label={ariaLabel}
        className="h-64 w-full"
      >
        <BarChart data={series} accessibilityLayer>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="period" tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} width={48} />
          <ChartTooltip
            content={
              <ChartTooltipContent
                labelFormatter={(label, payload) => {
                  const point = payload[0]?.payload as PenetrationPoint | undefined
                  if (point?.penetrationRate == null) return label
                  return `${label}, доля двух и более: ${formatPercent(point.penetrationRate)}`
                }}
              />
            }
          />
          <ChartLegend content={<ChartLegendContent />} itemSorter={null} />
          {segmentKeys.map((key) => (
            <Bar key={key} dataKey={key} stackId="guests" fill={`var(--color-${key})`} />
          ))}
        </BarChart>
      </ChartContainer>
      <figcaption className="text-xs text-muted-foreground">
        Распределение гостей по числу арендаторов; в подсказке — доля гостей с двумя и более
      </figcaption>
    </figure>
  )
}
