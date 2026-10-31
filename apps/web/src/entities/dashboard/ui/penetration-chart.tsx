import { formatPercent } from '@/shared/lib/format'
import { EmptyState } from '@/shared/ui/empty-state'
import type { PenetrationPoint } from '../model/types'

const numberFormat = new Intl.NumberFormat('ru-RU')

const segments = [
  { key: 'one', label: '1 арендатор', className: 'bg-chart-3' },
  { key: 'two', label: '2 арендатора', className: 'bg-chart-2' },
  { key: 'threePlus', label: '3 и более', className: 'bg-chart-1' },
] as const

interface PenetrationChartProps {
  series: PenetrationPoint[]
}

export function PenetrationChart({ series }: PenetrationChartProps) {
  if (series.length === 0) return <EmptyState title="Нет данных за период" />

  return (
    <figure className="flex flex-col gap-3">
      <ul className="flex flex-wrap gap-4 text-xs">
        {segments.map((segment) => (
          <li key={segment.key} className="flex items-center gap-1.5">
            <span className={`size-3 rounded-sm ${segment.className}`} aria-hidden />
            {segment.label}
          </li>
        ))}
      </ul>
      <ul className="flex flex-col gap-2">
        {series.map((point) => {
          const total = point.one + point.two + point.threePlus
          const summary = segments
            .map((s) => `${s.label}: ${numberFormat.format(point[s.key])}`)
            .join(', ')
          return (
            <li
              key={point.period}
              className="grid grid-cols-[5rem_1fr_4rem] items-center gap-3 text-sm"
            >
              <span>{point.period}</span>
              <div
                role="img"
                aria-label={`${point.period}. ${summary}`}
                className="flex h-5 overflow-hidden rounded-sm bg-muted"
              >
                {segments.map((segment) => (
                  <div
                    key={segment.key}
                    className={segment.className}
                    style={{ width: `${total === 0 ? 0 : (point[segment.key] / total) * 100}%` }}
                  />
                ))}
              </div>
              <span className="text-right text-muted-foreground">
                {point.penetrationRate === undefined ? '—' : formatPercent(point.penetrationRate)}
              </span>
            </li>
          )
        })}
      </ul>
      <figcaption className="text-xs text-muted-foreground">
        Распределение гостей по числу арендаторов; справа — доля гостей с двумя и более
      </figcaption>
    </figure>
  )
}
