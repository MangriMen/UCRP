import type { Guest, GuestContact, GuestEvent } from '../model/types'

export const mockGuests: Guest[] = [
  {
    id: 'guest-0001',
    displayName: 'Иван Петров',
    status: 'active',
    tenantCount: 3,
    eventCount: 14,
    firstSeenAt: '2026-01-12T10:00:00Z',
    lastSeenAt: '2026-10-28T18:20:00Z',
    createdAt: '2026-01-12T10:00:00Z',
    updatedAt: '2026-10-28T18:20:00Z',
  },
  {
    id: 'guest-0002',
    displayName: 'Мария Иванова',
    status: 'active',
    tenantCount: 1,
    eventCount: 4,
    firstSeenAt: '2026-06-03T12:30:00Z',
    lastSeenAt: '2026-10-02T15:10:00Z',
    createdAt: '2026-06-03T12:30:00Z',
    updatedAt: '2026-10-02T15:10:00Z',
  },
  {
    id: 'guest-0003',
    displayName: null,
    status: 'active',
    tenantCount: 2,
    eventCount: 7,
    firstSeenAt: null,
    lastSeenAt: null,
    createdAt: '2026-08-19T09:00:00Z',
    updatedAt: '2026-09-01T09:00:00Z',
  },
  {
    id: 'guest-0004',
    displayName: 'Пётр Сидоров',
    status: 'merged',
    mergedIntoId: 'guest-0001',
    tenantCount: 1,
    eventCount: 2,
    firstSeenAt: '2026-03-01T08:00:00Z',
    lastSeenAt: '2026-03-15T08:00:00Z',
    createdAt: '2026-03-01T08:00:00Z',
    updatedAt: '2026-09-20T08:00:00Z',
  },
]

export const mockGuestEvents: GuestEvent[] = [
  {
    id: 'event-0001',
    guestId: 'guest-0001',
    tenantId: 'tenant-0001',
    tenantName: 'Лента',
    eventTypeId: 'type-purchase',
    eventTypeCode: 'purchase',
    occurredAt: '2026-10-28T18:20:00Z',
    amountMinor: 350000,
    currency: 'RUB',
    createdAt: '2026-10-28T18:30:00Z',
  },
  {
    id: 'event-0002',
    guestId: 'guest-0001',
    tenantId: 'tenant-0002',
    tenantName: 'Зара',
    eventTypeId: 'type-visit',
    eventTypeCode: 'visit',
    occurredAt: '2026-10-20T12:00:00Z',
    createdAt: '2026-10-20T12:05:00Z',
  },
]

export const mockGuestContacts: GuestContact[] = [
  {
    id: 'contact-0001',
    guestId: 'guest-0001',
    exportId: 'export-0001',
    channel: 'email',
    contactedAt: '2026-10-25T10:00:00Z',
  },
]
