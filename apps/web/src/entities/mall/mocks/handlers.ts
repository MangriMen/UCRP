import { http, HttpResponse } from 'msw'
import { mockTenants } from '@/entities/tenant/mocks'
import { API, apiError, mockId } from '@/shared/api/mocks/helpers'
import type { Mall, MallCreateRequest, MallUpdateRequest } from '../model/types'
import { mockMalls } from './fixtures'

const malls: Mall[] = structuredClone(mockMalls)

const isCodeTaken = (code: string) => malls.some((m) => m.code === code)

export const mallHandlers = [
  http.get(`${API}/malls`, () => HttpResponse.json(malls)),

  http.post(`${API}/malls`, async ({ request }) => {
    const body = (await request.json()) as MallCreateRequest
    if (isCodeTaken(body.code)) {
      return apiError(409, 'duplicate', 'ТЦ с таким кодом уже существует')
    }
    const mall: Mall = {
      id: mockId('mall-new'),
      code: body.code,
      name: body.name,
      isActive: true,
      floors: [],
    }
    malls.push(mall)
    return HttpResponse.json(mall, { status: 201 })
  }),

  http.patch(`${API}/malls/:id`, async ({ params, request }) => {
    const mall = malls.find((m) => m.id === params.id)
    if (!mall) return apiError(404, 'not_found', 'ТЦ не найден')
    const body = (await request.json()) as MallUpdateRequest
    Object.assign(mall, body)
    return HttpResponse.json(mall)
  }),

  http.delete(`${API}/malls/:id`, ({ params }) => {
    const index = malls.findIndex((m) => m.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'ТЦ не найден')
    if (mockTenants.some((t) => t.mallId === params.id)) {
      return apiError(409, 'conflict', 'У ТЦ есть арендаторы, удаление невозможно')
    }
    malls.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
