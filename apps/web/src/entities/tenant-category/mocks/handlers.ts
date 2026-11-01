import { http, HttpResponse } from 'msw'
import { API, apiError, mockId } from '@/shared/api/mocks/helpers'
import { DEFAULT_RELATION_WEIGHT_PCT } from '../model/relation'
import type {
  TenantCategory,
  TenantCategoryRelation,
  TenantCategoryRelations,
  TenantCategoryRequest,
} from '../model/types'
import { CATEGORY_IN_USE_ID, mockTenantCategories, mockTenantCategoryRelations } from './fixtures'

const categories: TenantCategory[] = structuredClone(mockTenantCategories)
let relations: TenantCategoryRelation[] = structuredClone(mockTenantCategoryRelations)

const isCodeTaken = (code: string, exceptId?: string) =>
  categories.some((c) => c.code === code && c.id !== exceptId)

const involves = (relation: TenantCategoryRelation, categoryId: string) =>
  relation.categoryId === categoryId || relation.relatedCategoryId === categoryId

function relationsOf(categoryId: string): TenantCategoryRelations {
  const items = relations.filter((r) => involves(r, categoryId))
  return {
    categoryId,
    relations: items.map((r) => {
      const relatedCategoryId = r.categoryId === categoryId ? r.relatedCategoryId : r.categoryId
      return {
        categoryId,
        relatedCategoryId,
        relatedCategoryName: categories.find((c) => c.id === relatedCategoryId)?.name,
        relationType: r.relationType,
        weightPct: r.weightPct,
      }
    }),
  }
}

function isValidRelationSet(categoryId: string, items: TenantCategoryRelation[]) {
  const related = items.map((r) => r.relatedCategoryId)
  return (
    new Set(related).size === related.length &&
    items.every(
      (r) =>
        r.categoryId === categoryId &&
        r.relatedCategoryId !== categoryId &&
        categories.some((c) => c.id === r.relatedCategoryId) &&
        (r.relationType === 'adjacent' || r.relationType === 'competitor') &&
        (r.weightPct == null || (r.weightPct >= 0 && r.weightPct <= 100)),
    )
  )
}

export const tenantCategoryHandlers = [
  http.get(`${API}/tenant-categories`, ({ request }) => {
    const includeInactive = new URL(request.url).searchParams.get('include_inactive') === 'true'
    return HttpResponse.json(includeInactive ? categories : categories.filter((c) => c.isActive))
  }),

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
      isActive: true,
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
    if (relations.some((r) => involves(r, String(params.id)))) {
      return apiError(409, 'category_in_use', 'Категория используется в связях категорий')
    }
    categories.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),

  http.get(`${API}/tenant-categories/:id/relations`, ({ params }) => {
    const categoryId = String(params.id)
    if (!categories.some((c) => c.id === categoryId)) {
      return apiError(404, 'not_found', 'Категория не найдена')
    }
    return HttpResponse.json(relationsOf(categoryId))
  }),

  http.put(`${API}/tenant-categories/:id/relations`, async ({ params, request }) => {
    const categoryId = String(params.id)
    if (!categories.some((c) => c.id === categoryId)) {
      return apiError(404, 'not_found', 'Категория не найдена')
    }
    const body = (await request.json()) as { relations: TenantCategoryRelation[] }
    if (!isValidRelationSet(categoryId, body.relations)) {
      return apiError(400, 'validation_error', 'Некорректные данные')
    }
    relations = [
      ...relations.filter((r) => !involves(r, categoryId)),
      ...body.relations.map((r) => ({
        categoryId,
        relatedCategoryId: r.relatedCategoryId,
        relationType: r.relationType,
        weightPct: r.weightPct ?? DEFAULT_RELATION_WEIGHT_PCT,
      })),
    ]
    return HttpResponse.json(relationsOf(categoryId))
  }),
]
