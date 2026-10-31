import { http, HttpResponse } from 'msw'
import { API, apiError, mockId } from '@/shared/api/mocks/helpers'
import type { EventType, EventTypeRequest } from '../model/types'
import { EVENT_TYPE_IN_USE_ID, mockEventTypes } from './fixtures'

const eventTypes: EventType[] = structuredClone(mockEventTypes)

const isCodeTaken = (code: string, exceptId?: string) =>
  eventTypes.some((t) => t.code === code && t.id !== exceptId)

export const eventTypeHandlers = [
  http.get(`${API}/event-types`, () => HttpResponse.json(eventTypes)),

  http.post(`${API}/event-types`, async ({ request }) => {
    const body = (await request.json()) as EventTypeRequest
    if (isCodeTaken(body.code)) {
      return apiError(409, 'duplicate', 'Тип события с таким кодом уже существует')
    }
    const eventType: EventType = {
      ...body,
      id: mockId('event-type-new'),
      isActive: body.isActive ?? true,
    }
    eventTypes.push(eventType)
    return HttpResponse.json(eventType, { status: 201 })
  }),

  http.patch(`${API}/event-types/:id`, async ({ params, request }) => {
    const eventType = eventTypes.find((t) => t.id === params.id)
    if (!eventType) return apiError(404, 'not_found', 'Тип события не найден')
    const body = (await request.json()) as EventTypeRequest
    if (isCodeTaken(body.code, eventType.id)) {
      return apiError(409, 'duplicate', 'Тип события с таким кодом уже существует')
    }
    Object.assign(eventType, body)
    return HttpResponse.json(eventType)
  }),

  http.delete(`${API}/event-types/:id`, ({ params }) => {
    const index = eventTypes.findIndex((t) => t.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'Тип события не найден')
    if (params.id === EVENT_TYPE_IN_USE_ID) {
      return apiError(409, 'event_type_in_use', 'Тип события используется в событиях')
    }
    eventTypes.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
