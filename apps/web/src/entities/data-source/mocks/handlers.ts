import { http, HttpResponse } from 'msw'
import { mockDataSourceTypes } from '@/entities/data-source-type/mocks'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type { DataSource, DataSourceCreateRequest, DataSourceUpdateRequest } from '../model/types'
import { DATA_SOURCE_WITH_IMPORTS_ID, mockDataSources } from './fixtures'

const sources: DataSource[] = structuredClone(mockDataSources)

export const dataSourceHandlers = [
  http.get(`${API}/data-sources`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const tenantId = params.get('tenant_id')
    const typeId = params.get('type_id')
    const isActive = params.get('is_active')
    const filtered = sources.filter(
      (s) =>
        (!tenantId || s.tenantId === tenantId) &&
        (!typeId || s.typeId === typeId) &&
        (isActive === null || String(s.isActive) === isActive),
    )
    return HttpResponse.json(paged(request, filtered))
  }),

  http.get(`${API}/data-sources/:id`, ({ params }) => {
    const source = sources.find((s) => s.id === params.id)
    if (!source) return apiError(404, 'not_found', 'Источник не найден')
    return HttpResponse.json(source)
  }),

  http.post(`${API}/data-sources`, async ({ request }) => {
    const body = (await request.json()) as DataSourceCreateRequest
    const type = mockDataSourceTypes.find((t) => t.id === body.typeId)
    if (!type) return apiError(400, 'validation_error', 'Неизвестный тип источника')
    const now = new Date().toISOString()
    const source: DataSource = {
      ...body,
      id: mockId('ds-new'),
      typeCode: type.code,
      tenantName: null,
      isActive: body.isActive ?? true,
      lastImportAt: null,
      createdAt: now,
      updatedAt: now,
    }
    sources.push(source)
    return HttpResponse.json(source, { status: 201 })
  }),

  http.patch(`${API}/data-sources/:id`, async ({ params, request }) => {
    const source = sources.find((s) => s.id === params.id)
    if (!source) return apiError(404, 'not_found', 'Источник не найден')
    const body = (await request.json()) as DataSourceUpdateRequest
    Object.assign(source, body, { updatedAt: new Date().toISOString() })
    return HttpResponse.json(source)
  }),

  http.delete(`${API}/data-sources/:id`, ({ params }) => {
    const index = sources.findIndex((s) => s.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'Источник не найден')
    if (params.id === DATA_SOURCE_WITH_IMPORTS_ID) {
      return apiError(409, 'conflict', 'У источника есть импорты, удаление невозможно')
    }
    sources.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
