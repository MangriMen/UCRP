import { http, HttpResponse } from 'msw'
import { API, paged } from '@/shared/api/mocks/helpers'
import { mockDataSources } from './fixtures'

export const dataSourceHandlers = [
  http.get(`${API}/data-sources`, ({ request }) =>
    HttpResponse.json(paged(request, mockDataSources)),
  ),
]
