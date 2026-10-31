import { http, HttpResponse } from 'msw'
import { API, apiError, mockId } from '@/shared/api/mocks/helpers'
import type { DataSourceType, DataSourceTypeRequest } from '../model/types'
import { mockDataSourceTypes } from './fixtures'

const types: DataSourceType[] = structuredClone(mockDataSourceTypes)

const isCodeTaken = (code: string, exceptId?: string) =>
  types.some((t) => t.code === code && t.id !== exceptId)

export const dataSourceTypeHandlers = [
  http.get(`${API}/data-source-types`, () => HttpResponse.json(types)),

  http.post(`${API}/data-source-types`, async ({ request }) => {
    const body = (await request.json()) as DataSourceTypeRequest
    if (isCodeTaken(body.code)) {
      return apiError(409, 'duplicate', 'Тип источника с таким кодом уже существует')
    }
    const type: DataSourceType = {
      ...body,
      id: mockId('ds-type-new'),
      isActive: body.isActive ?? true,
    }
    types.push(type)
    return HttpResponse.json(type, { status: 201 })
  }),

  http.patch(`${API}/data-source-types/:id`, async ({ params, request }) => {
    const type = types.find((t) => t.id === params.id)
    if (!type) return apiError(404, 'not_found', 'Тип источника не найден')
    const body = (await request.json()) as DataSourceTypeRequest
    if (isCodeTaken(body.code, type.id)) {
      return apiError(409, 'duplicate', 'Тип источника с таким кодом уже существует')
    }
    Object.assign(type, body)
    return HttpResponse.json(type)
  }),
]
