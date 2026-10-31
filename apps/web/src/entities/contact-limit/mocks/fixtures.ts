import type { ContactLimit } from '../model/types'

export const mockContactLimits: ContactLimit[] = [
  {
    id: 'limit-0001',
    channel: null,
    maxContacts: 4,
    periodDays: 30,
    createdAt: '2026-09-01T09:00:00Z',
    updatedAt: '2026-09-01T09:00:00Z',
  },
  {
    id: 'limit-0002',
    channel: 'sms',
    maxContacts: 2,
    periodDays: 7,
    createdAt: '2026-09-10T09:00:00Z',
    updatedAt: '2026-09-10T09:00:00Z',
  },
]
