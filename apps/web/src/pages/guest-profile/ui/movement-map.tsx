import { ArrowDown, Building2 } from 'lucide-react'
import { useMemo } from 'react'
import { useGuestMovement } from '@/entities/guest'
import { formatDate, formatTime } from '@/shared/lib/format'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Skeleton } from '@/shared/ui/skeleton'
import { groupMovement, type MovementLeg } from '../lib/group-movement'

const floorLabel = (floor: number | null) => (floor === null ? 'Без этажа' : `${floor} этаж`)

const timeRange = (leg: MovementLeg) =>
  leg.startedAt === leg.endedAt
    ? `${formatDate(leg.startedAt)}, ${formatTime(leg.startedAt)}`
    : `${formatDate(leg.startedAt)}, ${formatTime(leg.startedAt)} – ${formatTime(leg.endedAt)}`

function LegCard({ leg }: { leg: MovementLeg }) {
  return (
    <div className="flex flex-col gap-1 rounded-md border p-3">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Building2 aria-hidden className="size-4 shrink-0 text-muted-foreground" />
        <span>{leg.mallName}</span>
        <span className="text-muted-foreground">·</span>
        <span>{floorLabel(leg.floor)}</span>
      </div>
      <div className="text-xs text-muted-foreground">{timeRange(leg)}</div>
      <p className="text-sm">{leg.tenantNames.join(', ')}</p>
      {leg.purchaseCount > 0 && (
        <div className="text-xs text-muted-foreground">Покупок: {leg.purchaseCount}</div>
      )}
    </div>
  )
}

export function MovementMap({ guestId }: { guestId: string }) {
  const movement = useGuestMovement(guestId)
  const items = movement.data?.items
  const legs = useMemo(() => (items ? groupMovement(items) : []), [items])

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">Карта передвижения</h2>
      {movement.isLoading && <Skeleton className="h-32 w-full" />}
      {movement.error && <ErrorBanner error={movement.error} />}
      {movement.data && legs.length === 0 && <EmptyState title="Передвижений пока нет" />}
      {legs.length > 0 && (
        <ol aria-label="Маршрут по этажам" className="flex flex-col items-stretch gap-1">
          {legs.map((leg, index) => (
            <li key={leg.key} className="flex flex-col items-center gap-1">
              {index > 0 && (
                <ArrowDown aria-hidden className="size-4 shrink-0 text-muted-foreground" />
              )}
              <div className="w-full">
                <LegCard leg={leg} />
              </div>
            </li>
          ))}
        </ol>
      )}
    </section>
  )
}
