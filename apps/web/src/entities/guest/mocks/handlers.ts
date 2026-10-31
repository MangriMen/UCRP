import { http, HttpResponse } from 'msw'
import { API, apiError, paged } from '@/shared/api/mocks/helpers'
import { MIN_SEARCH_LENGTH } from '../model/keys'
import {
  emptyProfile,
  mockGuestContacts,
  mockGuestEvents,
  mockGuestProfileExtras,
  mockGuests,
} from './fixtures'

const digits = (value: string) => value.replace(/\D/g, '')

export const guestHandlers = [
  http.get(`${API}/guests`, ({ request }) => {
    const query = new URL(request.url).searchParams.get('q')?.trim().toLowerCase() ?? ''
    if (query.length < MIN_SEARCH_LENGTH) {
      return apiError(400, 'validation_error', 'Запрос слишком короткий')
    }
    const matched = mockGuests.filter(
      (guest) =>
        guest.status !== 'merged' &&
        ((guest.displayName?.toLowerCase().includes(query) ?? false) ||
          (digits(query) !== '' && guest.id.includes(digits(query)))),
    )
    return HttpResponse.json(paged(request, matched))
  }),

  http.get(`${API}/guests/:id`, ({ params }) => {
    const guest = mockGuests.find((g) => g.id === params.id)
    if (!guest) return apiError(404, 'not_found', 'Гость не найден')
    return HttpResponse.json({ ...emptyProfile, ...guest, ...mockGuestProfileExtras[guest.id] })
  }),

  http.get(`${API}/guests/:id/events`, ({ params, request }) => {
    const tenantId = new URL(request.url).searchParams.get('tenant_id')
    const events = mockGuestEvents.filter(
      (e) => e.guestId === params.id && (!tenantId || e.tenantId === tenantId),
    )
    return HttpResponse.json(paged(request, events))
  }),

  http.get(`${API}/guests/:id/contacts`, ({ params, request }) =>
    HttpResponse.json(
      paged(
        request,
        mockGuestContacts.filter((c) => c.guestId === params.id),
      ),
    ),
  ),
]
