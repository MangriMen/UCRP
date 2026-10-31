import type { GuestEvent } from '@/entities/guest'
import { dayKey } from './day-key'

export interface EventDayGroup {
  day: string
  events: GuestEvent[]
}

export function groupEventsByDay(events: GuestEvent[]): EventDayGroup[] {
  const sorted = [...events].sort((a, b) => b.occurredAt.localeCompare(a.occurredAt))
  const groups: EventDayGroup[] = []
  for (const event of sorted) {
    const last = groups.at(-1)
    if (last && dayKey(last.events[0].occurredAt) === dayKey(event.occurredAt)) {
      last.events.push(event)
    } else {
      groups.push({ day: event.occurredAt, events: [event] })
    }
  }
  return groups
}
