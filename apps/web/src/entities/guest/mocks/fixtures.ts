import type {
  Guest,
  GuestContact,
  GuestEvent,
  GuestMovementStep,
  GuestProfile,
  MaskedGuestIdentifier,
} from '../model/types'

export const mockGuests: Guest[] = [
  {
    id: 'guest-0001',
    displayName: 'Иван Петров',
    status: 'active',
    tenantCount: 3,
    eventCount: 14,
    purchaseCount: 3,
    totalAmountMinor: 1250000,
    currency: 'RUB',
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
    purchaseCount: 0,
    totalAmountMinor: 0,
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
    purchaseCount: 0,
    totalAmountMinor: 0,
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
    purchaseCount: 0,
    totalAmountMinor: 0,
    firstSeenAt: '2026-03-01T08:00:00Z',
    lastSeenAt: '2026-03-15T08:00:00Z',
    createdAt: '2026-03-01T08:00:00Z',
    updatedAt: '2026-09-20T08:00:00Z',
  },
]

export const mockGuestMaskedIdentifiers: Record<string, MaskedGuestIdentifier[]> = {
  'guest-0001': [
    { idType: 'phone', valueMasked: '***4567' },
    { idType: 'email', valueMasked: 'i***@example.com' },
    { idType: 'card', valueMasked: '***1234' },
  ],
  'guest-0002': [{ idType: 'phone', valueMasked: '***8899' }],
  'guest-0004': [{ idType: 'loyalty_id', valueMasked: '***0042' }],
}

export const mockGuestSegmentMembers: Record<string, string[]> = {
  'segment-0001': ['guest-0001', 'guest-0002'],
}

export const mockGuestEvents: GuestEvent[] = [
  {
    id: 'event-0001',
    guestId: 'guest-0001',
    tenantId: 'tenant-0001',
    tenantName: 'Лента',
    dataSourceId: 'ds-0001',
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
    dataSourceId: 'ds-0002',
    eventTypeId: 'type-visit',
    eventTypeCode: 'visit',
    occurredAt: '2026-10-20T12:00:00Z',
    createdAt: '2026-10-20T12:05:00Z',
  },
  {
    id: 'event-0003',
    guestId: 'guest-0002',
    tenantId: 'tenant-0001',
    tenantName: 'Лента',
    dataSourceId: 'ds-0001',
    eventTypeId: 'type-visit',
    eventTypeCode: 'visit',
    occurredAt: '2026-10-02T15:10:00Z',
    createdAt: '2026-10-02T15:15:00Z',
  },
  {
    id: 'event-0004',
    guestId: 'guest-0003',
    tenantId: 'tenant-0002',
    tenantName: 'Зара',
    dataSourceId: 'ds-0002',
    eventTypeId: 'type-visit',
    eventTypeCode: 'visit',
    occurredAt: '2026-08-19T09:00:00Z',
    createdAt: '2026-08-19T09:05:00Z',
  },
]

const centralMall = { mallId: 'mall-0001', mallName: 'Центральный ТЦ' }

export const mockGuestMovement: Record<string, GuestMovementStep[]> = {
  'guest-0001': [
    {
      ...centralMall,
      eventId: 'event-0002',
      eventTypeCode: 'visit',
      tenantId: 'tenant-0002',
      tenantName: 'Зара',
      floor: 2,
      occurredAt: '2026-10-20T12:00:00Z',
    },
    {
      ...centralMall,
      eventId: 'event-0005',
      eventTypeCode: 'purchase',
      tenantId: 'tenant-0003',
      tenantName: 'Киномир',
      floor: 4,
      occurredAt: '2026-10-20T12:40:00Z',
    },
    {
      ...centralMall,
      eventId: 'event-0006',
      eventTypeCode: 'visit',
      tenantId: 'tenant-0003',
      tenantName: 'Киномир',
      floor: 4,
      occurredAt: '2026-10-20T13:05:00Z',
    },
    {
      ...centralMall,
      eventId: 'event-0001',
      eventTypeCode: 'purchase',
      tenantId: 'tenant-0001',
      tenantName: 'Лента',
      floor: 1,
      occurredAt: '2026-10-28T18:20:00Z',
    },
    {
      mallId: 'mall-0002',
      mallName: 'Северный ТЦ',
      eventId: 'event-0007',
      eventTypeCode: 'visit',
      tenantId: 'tenant-0101',
      tenantName: 'ФитЗона',
      floor: 2,
      occurredAt: '2026-10-28T19:30:00Z',
    },
  ],
}

export const mockGuestContacts: GuestContact[] = [
  {
    id: 'contact-0001',
    guestId: 'guest-0001',
    exportId: 'export-0001',
    channel: 'email',
    contactedAt: '2026-10-25T10:00:00Z',
  },
]

type ProfileExtras = Pick<GuestProfile, 'identifiers' | 'tenants' | 'segments'> &
  Partial<Pick<GuestProfile, 'purchaseCount' | 'totalAmountMinor' | 'currency'>>

export const emptyProfile: ProfileExtras = { identifiers: [], tenants: [], segments: [] }

export const mockGuestProfileExtras: Record<string, ProfileExtras> = {
  'guest-0001': {
    identifiers: [
      {
        id: 'identifier-0001',
        guestId: 'guest-0001',
        idType: 'phone',
        valueNormalized: '79001234567',
        isReliable: true,
      },
      {
        id: 'identifier-0002',
        guestId: 'guest-0001',
        idType: 'email',
        valueNormalized: 'ivan.petrov@example.com',
        isReliable: true,
      },
    ],
    tenants: [
      {
        tenantId: 'tenant-0001',
        tenantName: 'Лента',
        categoryName: 'Продукты',
        floor: 1,
        eventCount: 9,
        purchaseCount: 3,
        totalAmountMinor: 950000,
        currency: 'RUB',
        lastVisitAt: '2026-10-28T18:20:00Z',
      },
      {
        tenantId: 'tenant-0002',
        tenantName: 'Зара',
        categoryName: 'Одежда',
        floor: 2,
        eventCount: 5,
        purchaseCount: 1,
        totalAmountMinor: 300000,
        currency: 'RUB',
        lastVisitAt: '2026-10-20T12:00:00Z',
      },
    ],
    segments: [{ segmentId: 'segment-0001', name: 'Постоянные гости', versionNo: 2 }],
    purchaseCount: 4,
    totalAmountMinor: 1250000,
    currency: 'RUB',
  },
}
