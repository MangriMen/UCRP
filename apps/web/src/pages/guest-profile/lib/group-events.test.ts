import { describe, expect, it } from 'vitest'
import type { GuestEvent } from '@/entities/guest'
import { groupEventsByDay } from './group-events'

const event = (id: string, occurredAt: string): GuestEvent => ({
  id,
  tenantId: 'tenant-0001',
  eventTypeId: 'type-visit',
  occurredAt,
})

describe('groupEventsByDay', () => {
  it('groups events of one day and orders days from newest', () => {
    const groups = groupEventsByDay([
      event('a', '2026-10-20T09:00:00'),
      event('b', '2026-10-28T09:00:00'),
      event('c', '2026-10-28T18:00:00'),
    ])
    expect(groups.map((g) => g.events.map((e) => e.id))).toEqual([['c', 'b'], ['a']])
  })

  it('returns no groups for no events', () => {
    expect(groupEventsByDay([])).toEqual([])
  })
})
