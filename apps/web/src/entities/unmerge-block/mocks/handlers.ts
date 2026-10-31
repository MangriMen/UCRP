import { http, HttpResponse } from 'msw'
import { API, apiError, paged } from '@/shared/api/mocks/helpers'
import { mockUnmergeBlocks } from './fixtures'

export const unmergeBlockHandlers = [
  http.get(`${API}/unmerge-blocks`, ({ request }) =>
    HttpResponse.json(paged(request, mockUnmergeBlocks)),
  ),

  http.delete(`${API}/unmerge-blocks/:id`, ({ params }) => {
    if (!mockUnmergeBlocks.some((b) => b.id === params.id)) {
      return apiError(404, 'not_found', 'Запрет не найден')
    }
    return new HttpResponse(null, { status: 204 })
  }),
]
