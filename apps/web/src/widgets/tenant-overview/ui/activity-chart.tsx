import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts'
import type { TenantAggregateDay } from '@/entities/tenant-aggregates'
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
} satisfies ChartConfig

interface ActivityChartProps {
  series: TenantAggregateDay[]
  showNewGuests: boolean
}

export function ActivityChart({ series, showNewGuests }: ActivityChartProps) {
  return (
    <figure className="flex flex-col gap-3">
      <ChartContainer
        config={chartConfig}
        role="img"
        aria-label={`Динамика по дням: ${series.length} дн.`}
        className="h-64 w-full"
      >
        <BarChart data={series} accessibilityLayer>
          <CartesianGrid vertical={false} />
          <XAxis dataKey="date" tickLine={false} axisLine={false} minTickGap={24} />
          <YAxis tickLine={false} axisLine={false} width={40} allowDecimals={false} />
          <ChartTooltip content={<ChartTooltipContent />} />
          <ChartLegend content={<ChartLegendContent />} />
          <Bar dataKey="visitsCount" fill="var(--color-visitsCount)" radius={2} />
          {showNewGuests && (
            <Bar dataKey="newGuestsCount" fill="var(--color-newGuestsCount)" radius={2} />
          )}
        </BarChart>
      </ChartContainer>
      <figcaption className="text-xs text-muted-foreground">
        Дни, где гостей меньше порога, не показываются
      </figcaption>
    </figure>
  )
}
