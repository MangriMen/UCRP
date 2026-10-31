import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { mockQualityDetails, mockQualitySources } from './fixtures'

export const dataQualityHandlers = [
  http.get(`${API}/data-quality/sources`, () =>
    HttpResponse.json({
      items: mockQualitySources.map((source) => ({
        dataSourceId: source.dataSourceId,
        name: source.dataSourceName,
        lastImportAt: source.lastSuccessfulImportAt ?? undefined,
        rejectedRate: source.rejectedShare30d,
        pendingMerges: source.pendingMergeCandidates,
        warning: source.hasWarning,
        suppressed: false,
      })),
      meta: { page: 1, pageSize: 50, total: mockQualitySources.length, totalPages: 1 },
    }),
  ),

  http.get(`${API}/data-quality/sources/:id`, ({ params }) => {
    const source = mockQualitySources.find((s) => s.dataSourceId === params.id)
    if (!source) return apiError(404, 'not_found', 'Источник не найден')
    return HttpResponse.json({ ...source, ...mockQualityDetails[source.dataSourceId] })
  }),
]
