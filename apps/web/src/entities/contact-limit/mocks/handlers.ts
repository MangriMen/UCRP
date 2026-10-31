import { http, HttpResponse } from 'msw'
import { API, apiError, mockId } from '@/shared/api/mocks/helpers'
import type {
  ContactLimit,
  ContactLimitCreateRequest,
  ContactLimitUpdateRequest,
} from '../model/types'
import { mockContactLimits } from './fixtures'

const limits: ContactLimit[] = structuredClone(mockContactLimits)

export const contactLimitHandlers = [
  http.get(`${API}/contact-limits`, () => HttpResponse.json(limits)),

  http.post(`${API}/contact-limits`, async ({ request }) => {
    const body = (await request.json()) as ContactLimitCreateRequest
    const now = new Date().toISOString()
    const limit: ContactLimit = {
      id: mockId('limit-new'),
      channel: body.channel ?? null,
      maxContacts: body.maxContacts,
      periodDays: body.periodDays,
      createdAt: now,
      updatedAt: now,
    }
    limits.push(limit)
    return HttpResponse.json(limit, { status: 201 })
  }),

  http.patch(`${API}/contact-limits/:id`, async ({ params, request }) => {
    const limit = limits.find((l) => l.id === params.id)
    if (!limit) return apiError(404, 'not_found', 'Лимит не найден')
    const body = (await request.json()) as ContactLimitUpdateRequest
    Object.assign(limit, body, { updatedAt: new Date().toISOString() })
    return HttpResponse.json(limit)
  }),

  http.delete(`${API}/contact-limits/:id`, ({ params }) => {
    const index = limits.findIndex((l) => l.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'Лимит не найден')
    limits.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
