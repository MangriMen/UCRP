import { http, HttpResponse } from 'msw'
import { API, apiError, mockId } from '@/shared/api/mocks/helpers'
import type { TenantCategory, TenantCategoryRequest } from '../model/types'
import { CATEGORY_IN_USE_ID, mockTenantCategories } from './fixtures'

const categories: TenantCategory[] = structuredClone(mockTenantCategories)

const isCodeTaken = (code: string, exceptId?: string) =>
  categories.some((c) => c.code === code && c.id !== exceptId)

export const tenantCategoryHandlers = [
  http.get(`${API}/tenant-categories`, () => HttpResponse.json(categories)),

  http.post(`${API}/tenant-categories`, async ({ request }) => {
    const body = (await request.json()) as TenantCategoryRequest
    if (isCodeTaken(body.code)) {
      return apiError(409, 'duplicate', 'Категория с таким кодом уже существует')
    }
    const now = new Date().toISOString()
    const category: TenantCategory = {
      id: mockId('category-new'),
      code: body.code,
      name: body.name,
      description: body.description ?? null,
      createdAt: now,
      updatedAt: now,
    }
    categories.push(category)
    return HttpResponse.json(category, { status: 201 })
  }),

  http.patch(`${API}/tenant-categories/:id`, async ({ params, request }) => {
    const category = categories.find((c) => c.id === params.id)
    if (!category) return apiError(404, 'not_found', 'Категория не найдена')
    const body = (await request.json()) as TenantCategoryRequest
    if (isCodeTaken(body.code, category.id)) {
      return apiError(409, 'duplicate', 'Категория с таким кодом уже существует')
    }
    Object.assign(category, body, { updatedAt: new Date().toISOString() })
    return HttpResponse.json(category)
  }),

  http.delete(`${API}/tenant-categories/:id`, ({ params }) => {
    const index = categories.findIndex((c) => c.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'Категория не найдена')
    if (params.id === CATEGORY_IN_USE_ID) {
      return apiError(409, 'category_in_use', 'Категория используется арендаторами')
    }
    categories.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
