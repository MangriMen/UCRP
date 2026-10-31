import { http, HttpResponse } from 'msw'
import { API, apiError, paged } from '@/shared/api/mocks/helpers'
import { MIN_SEARCH_LENGTH } from '../model/keys'
import {
  emptyProfile,
  mockGuestContacts,
  mockGuestEvents,
  mockGuestMovement,
  mockGuestProfileExtras,
  mockGuests,
} from './fixtures'
import type { Guest } from '../model/types'

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

export const guestHandlers = [
  http.get(`${API}/guests`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const query = params.get('q')?.trim().toLowerCase() ?? ''
    if (params.has('q') && query.length < MIN_SEARCH_LENGTH) {
      return apiError(400, 'validation_error', 'Запрос слишком короткий')
    }
    const matched = mockGuests.filter(
      (guest) =>
        guest.status !== 'merged' &&
        matchesQuery(guest, query) &&
        hasMatchingEvent(guest.id, {
          sourceId: params.get('source_id'),
          tenantId: params.get('tenant_id'),
          periodFrom: params.get('period_from'),
          periodTo: params.get('period_to'),
        }),
    )
    return HttpResponse.json(paged(request, matched))
  }),

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
