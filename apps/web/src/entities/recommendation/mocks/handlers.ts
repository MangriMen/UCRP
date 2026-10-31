import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import type { RecommendationList } from '../model/types'
import { mockGuestRecommendations, mockSegmentRecommendations } from './fixtures'

const guestExists = (id: unknown) => typeof id === 'string' && id.startsWith('guest-')

const listFor = (guestId: string, computedAt?: string): RecommendationList => {
  const items = mockGuestRecommendations[guestId] ?? []
  return {
    guestId,
    computedAt: items[0]?.computedAt ?? null,
    items: computedAt ? items.map((item) => ({ ...item, computedAt })) : items,
  }
}

export const recommendationHandlers = [
  http.get(`${API}/guests/:id/recommendations`, ({ params }) => {
    if (!guestExists(params.id)) return apiError(404, 'not_found', 'Гость не найден')
    return HttpResponse.json(listFor(params.id as string))
  }),

  http.post(`${API}/guests/:id/recommendations/refresh`, ({ params }) => {
    if (!guestExists(params.id)) return apiError(404, 'not_found', 'Гость не найден')
    const now = new Date().toISOString()
    return HttpResponse.json({ ...listFor(params.id as string, now), computedAt: now })
  }),

  http.get(`${API}/segments/:id/recommendations`, ({ params, request }) => {
    if (typeof params.id !== 'string' || !params.id.startsWith('segment-')) {
      return apiError(404, 'not_found', 'Сегмент не найден')
    }
    const limit = Number(new URL(request.url).searchParams.get('limit') ?? 10)
    return HttpResponse.json({
      segmentId: params.id,
      items: mockSegmentRecommendations.slice(0, limit),
    })
  }),
]
