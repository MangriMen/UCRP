import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { mockQualityDetails, mockQualitySources } from './fixtures'

export const dataQualityHandlers = [
  http.get(`${API}/data-quality/sources`, () => HttpResponse.json(mockQualitySources)),

  http.get(`${API}/data-quality/sources/:id`, ({ params }) => {
    const source = mockQualitySources.find((s) => s.dataSourceId === params.id)
    if (!source) return apiError(404, 'not_found', 'Источник не найден')
    return HttpResponse.json({ ...source, ...mockQualityDetails[source.dataSourceId] })
  }),
]
