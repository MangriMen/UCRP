import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/shared/ui/chart'

const chartConfig = {
  visitsCount: { label: 'Визиты', color: 'var(--chart-1)' },
  newGuestsCount: { label: 'Новые гости', color: 'var(--chart-2)' },
  repeatVisitsCount: { label: 'Повторные визиты', color: 'var(--chart-3)' },
  revenue: { label: 'Выручка, ₽', color: 'var(--chart-4)' },
} satisfies ChartConfig

export type ActivityMetric = keyof typeof chartConfig

export type ActivityPoint = { date: string } & Partial<Record<ActivityMetric, number | null>>

interface ActivityChartProps {
  title: string
  points: ActivityPoint[]
  metrics: ActivityMetric[]
}

export function ActivityChart({ title, points, metrics }: ActivityChartProps) {
  return (
    <figure className="flex flex-col gap-3">
      <ChartContainer
        config={chartConfig}
        role="img"
        aria-label={`${title}: ${points.length} дн.`}
        className="h-64 w-full"
      >
        <BarChart data={points} accessibilityLayer>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="date" tickLine={false} axisLine={false} minTickGap={24} />
          <YAxis tickLine={false} axisLine={false} width={56} allowDecimals={false} />
          <ChartTooltip content={<ChartTooltipContent />} />
          <ChartLegend content={<ChartLegendContent />} />
          {metrics.map((metric) => (
            <Bar key={metric} dataKey={metric} fill={`var(--color-${metric})`} radius={2} />
          ))}
        </BarChart>
      </ChartContainer>
      <figcaption className="text-xs text-muted-foreground">
        Дни, где гостей меньше порога, не показываются
      </figcaption>
    </figure>
  )
}
