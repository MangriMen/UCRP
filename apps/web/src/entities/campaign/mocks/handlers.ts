import { http, HttpResponse } from 'msw'
import { API, apiError, paged } from '@/shared/api/mocks/helpers'
import { mockCampaigns, mockCampaignResults } from './fixtures'

export const campaignHandlers = [
  http.get(`${API}/campaigns`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const status = params.get('status')
    const segmentId = params.get('segment_id')
    const filtered = mockCampaigns.filter(
      (c) => (!status || c.status === status) && (!segmentId || c.segmentId === segmentId),
    )
    return HttpResponse.json(paged(request, filtered))
  }),

  http.get(`${API}/campaigns/:id`, ({ params }) => {
    const campaign = mockCampaigns.find((c) => c.id === params.id)
    if (!campaign) return apiError(404, 'not_found', 'Кампания не найдена')
    return HttpResponse.json(campaign)
  }),

  http.get(`${API}/campaigns/:id/results`, ({ params }) => {
    const results = mockCampaignResults[String(params.id)]
    if (!results) return apiError(404, 'not_found', 'Результаты кампании ещё не рассчитаны')
    return HttpResponse.json(results)
  }),
]
