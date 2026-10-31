import { http, HttpResponse } from 'msw'
import { API, apiError, paged } from '@/shared/api/mocks/helpers'
import { MIN_SEARCH_LENGTH } from '../model/keys'
import {
  emptyProfile,
  mockGuestContacts,
  mockGuestEvents,
  mockGuestMaskedIdentifiers,
  mockGuestMovement,
  mockGuestProfileExtras,
  mockGuestSegmentMembers,
  mockGuests,
} from './fixtures'
import type { Guest, GuestListItem } from '../model/types'

const digits = (value: string) => value.replace(/\D/g, '')

interface EventFilter {
  sourceId: string | null
  tenantId: string | null
  periodFrom: string | null
  periodTo: string | null
}

const matchesQuery = (guest: Guest, query: string) =>
  query === '' ||
  (guest.displayName?.toLowerCase().includes(query) ?? false) ||
  (digits(query) !== '' && guest.id.includes(digits(query)))

const hasMatchingEvent = (guestId: string, filter: EventFilter) => {
  const { sourceId, tenantId, periodFrom, periodTo } = filter
  if (!sourceId && !tenantId && !periodFrom && !periodTo) return true
  return mockGuestEvents.some((event) => {
    const day = event.occurredAt.slice(0, 10)
    return (
      event.guestId === guestId &&
      (!sourceId || event.dataSourceId === sourceId) &&
      (!tenantId || event.tenantId === tenantId) &&
      (!periodFrom || day >= periodFrom) &&
      (!periodTo || day <= periodTo)
    )
  })
}

const SORT_VALUES: Record<string, (guest: Guest) => number | string> = {
  createdAt: (guest) => guest.createdAt,
  eventCount: (guest) => guest.eventCount ?? 0,
  totalAmountMinor: (guest) => guest.totalAmountMinor ?? 0,
  firstSeenAt: (guest) => guest.firstSeenAt ?? '',
  lastSeenAt: (guest) => guest.lastSeenAt ?? '',
}

const compareBySort = (sort: string) => {
  const direction = sort.startsWith('-') ? -1 : 1
  const value = SORT_VALUES[sort.replace('-', '')] ?? SORT_VALUES.createdAt
  return (a: Guest, b: Guest) => {
    const left = value(a)
    const right = value(b)
    return left < right ? -direction : left > right ? direction : 0
  }
}

const matchesStatistics = (guest: Guest, params: URLSearchParams) => {
  const hasPurchases = params.get('has_purchases')
  const firstSeenFrom = params.get('first_seen_from')
  const lastSeenTo = params.get('last_seen_to')
  const segmentId = params.get('segment_id')
  const purchased = (guest.purchaseCount ?? 0) > 0
  return (
    (hasPurchases === null || (hasPurchases === 'true') === purchased) &&
    (!firstSeenFrom ||
      (guest.firstSeenAt != null && guest.firstSeenAt.slice(0, 10) >= firstSeenFrom)) &&
    (!lastSeenTo || (guest.lastSeenAt != null && guest.lastSeenAt.slice(0, 10) <= lastSeenTo)) &&
    (!segmentId || (mockGuestSegmentMembers[segmentId] ?? []).includes(guest.id))
  )
}

const listGuests = (request: Request, tenantId: string | null) => {
  const params = new URL(request.url).searchParams
  const query = params.get('q')?.trim().toLowerCase() ?? ''
  if (params.has('q') && query.length < MIN_SEARCH_LENGTH) {
    return apiError(400, 'validation_error', 'Запрос слишком короткий')
  }
  const status = params.get('status') ?? 'active'
  const matched = mockGuests
    .filter(
      (guest) =>
        guest.status === status &&
        matchesQuery(guest, query) &&
        matchesStatistics(guest, params) &&
        hasMatchingEvent(guest.id, {
          sourceId: params.get('source_id'),
          tenantId: tenantId ?? params.get('tenant_id'),
          periodFrom: params.get('period_from'),
          periodTo: params.get('period_to'),
        }),
    )
    .sort(compareBySort(params.get('sort') ?? '-createdAt'))
    .map((guest): GuestListItem => ({
      ...guest,
      maskedIdentifiers: mockGuestMaskedIdentifiers[guest.id] ?? [],
    }))
  return HttpResponse.json(paged(request, matched))
}

export const guestHandlers = [
  http.get(`${API}/guests`, ({ request }) => listGuests(request, null)),

  http.get(`${API}/tenants/:id/guests`, ({ params, request }) =>
    listGuests(request, String(params.id)),
  ),

  http.get(`${API}/guests/:id`, ({ params }) => {
    const guest = mockGuests.find((g) => g.id === params.id)
    if (!guest) return apiError(404, 'not_found', 'Гость не найден')
    return HttpResponse.json({ ...emptyProfile, ...guest, ...mockGuestProfileExtras[guest.id] })
  }),

  http.delete(`${API}/guests/:id`, ({ params }) =>
    mockGuests.some((g) => g.id === params.id)
      ? new HttpResponse(null, { status: 204 })
      : apiError(404, 'not_found', 'Гость не найден'),
  ),

  http.get(`${API}/guests/:id/events`, ({ params, request }) => {
    const query = new URL(request.url).searchParams
    const type = query.get('type')
    const tenantId = query.get('tenant_id')
    const from = query.get('from')
    const to = query.get('to')
    const amountMin = query.has('amount_min') ? Number(query.get('amount_min')) : null
    const amountMax = query.has('amount_max') ? Number(query.get('amount_max')) : null
    const sort = query.get('sort') ?? '-occurredAt'
    const direction = sort.startsWith('-') ? -1 : 1
    const field = sort.replace('-', '') === 'amountMinor' ? 'amountMinor' : 'occurredAt'
    const events = mockGuestEvents
      .filter(
        (e) =>
          e.guestId === params.id &&
          (!type || e.eventTypeCode === type) &&
          (!tenantId || e.tenantId === tenantId) &&
          (!from || e.occurredAt.slice(0, 10) >= from) &&
          (!to || e.occurredAt.slice(0, 10) <= to) &&
          (amountMin === null || (e.amountMinor ?? -1) >= amountMin) &&
          (amountMax === null || (e.amountMinor != null && e.amountMinor <= amountMax)),
      )
      .sort((a, b) => {
        const left = field === 'amountMinor' ? (a.amountMinor ?? -1) : a.occurredAt
        const right = field === 'amountMinor' ? (b.amountMinor ?? -1) : b.occurredAt
        return left < right ? -direction : left > right ? direction : 0
      })
    return HttpResponse.json(paged(request, events))
  }),

  http.get(`${API}/guests/:id/movement`, ({ params }) =>
    HttpResponse.json({ items: mockGuestMovement[String(params.id)] ?? [] }),
  ),

  http.get(`${API}/guests/:id/contacts`, ({ params, request }) =>
    HttpResponse.json(
      paged(
        request,
        mockGuestContacts.filter((c) => c.guestId === params.id),
      ),
    ),
  ),
]
