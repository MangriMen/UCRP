import { http, HttpResponse } from 'msw'
import { API, apiError, paged } from '@/shared/api/mocks/helpers'
import { mockMergeBlocks } from './fixtures'

export const mergeBlockHandlers = [
  http.get(`${API}/merge-blocks`, ({ request }) =>
    HttpResponse.json(paged(request, mockMergeBlocks)),
  ),

  http.delete(`${API}/merge-blocks/:id`, ({ params }) => {
    if (!mockMergeBlocks.some((b) => b.id === params.id)) {
      return apiError(404, 'not_found', 'Запрет не найден')
    }
    return new HttpResponse(null, { status: 204 })
  }),
]
