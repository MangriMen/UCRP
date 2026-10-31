import { CircleDot, MapPin, ShoppingBag, type LucideIcon } from 'lucide-react'
import { useState } from 'react'
import { useGuestEvents, type GuestEvent } from '@/entities/guest'
import { hasErrorCode } from '@/shared/api/errors'
import { formatDate, formatMoney, formatTime } from '@/shared/lib/format'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Skeleton } from '@/shared/ui/skeleton'
import { TablePagination } from '@/shared/ui/table-pagination'
import { groupEventsByDay } from '../lib/group-events'

const eventPresentation: Record<string, { label: string; icon: LucideIcon }> = {
  purchase: { label: 'Покупка', icon: ShoppingBag },
  visit: { label: 'Визит', icon: MapPin },
}

function EventRow({ event }: { event: GuestEvent }) {
  const code = event.eventTypeCode ?? ''
  const { label, icon: Icon } = eventPresentation[code] ?? {
    label: code || 'Событие',
    icon: CircleDot,
  }
  return (
    <li className="flex items-center gap-3 py-2">
      <Icon aria-hidden className="size-4 shrink-0 text-muted-foreground" />
      <div className="flex-1">
        <div className="text-sm font-medium">{label}</div>
        <div className="text-xs text-muted-foreground">{event.tenantName ?? event.tenantId}</div>
      </div>
      <div className="text-right text-sm">
        {event.amountMinor != null && <div>{formatMoney(event.amountMinor)}</div>}
        <div className="text-xs text-muted-foreground">{formatTime(event.occurredAt)}</div>
      </div>
    </li>
  )
}

export function EventFeed({ guestId }: { guestId: string }) {
  const [page, setPage] = useState(1)
  const events = useGuestEvents(guestId, { page })

  if (hasErrorCode(events.error, 'forbidden_field')) return null

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">Лента событий</h2>
      {events.isLoading && <Skeleton className="h-32 w-full" />}
      {events.error && <ErrorBanner error={events.error} />}
      {events.data?.items.length === 0 && <EmptyState title="Событий пока нет" />}
      {events.data &&
        groupEventsByDay(events.data.items).map((group) => (
          <div key={group.day}>
            <h3 className="text-sm font-medium text-muted-foreground">{formatDate(group.day)}</h3>
            <ul className="divide-y">
              {group.events.map((event) => (
                <EventRow key={event.id} event={event} />
              ))}
            </ul>
          </div>
        ))}
      {events.data && events.data.items.length > 0 && (
        <TablePagination meta={events.data.meta} onPageChange={setPage} />
      )}
    </section>
  )
}
