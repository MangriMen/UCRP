import { http, HttpResponse } from 'msw'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type { TenantCreateRequest, TenantListItem, TenantUpdateRequest } from '../model/types'
import { mockTenants, TENANT_WITH_EVENTS_ID } from './fixtures'

const tenants: TenantListItem[] = structuredClone(mockTenants)

const isCodeTaken = (code: string, exceptId?: string) =>
  tenants.some((t) => t.code.toLowerCase() === code.toLowerCase() && t.id !== exceptId)

const SORT_VALUES: Record<string, (tenant: TenantListItem) => number | string | null> = {
  name: (t) => t.name,
  floor: (t) => t.floor ?? null,
  priority: (t) => t.priority,
  createdAt: (t) => t.createdAt,
  guestsCount: (t) => t.guestsCount ?? null,
  visitsCount: (t) => t.visitsCount ?? null,
  newGuestsCount: (t) => t.newGuestsCount ?? null,
}

const compareBySort = (sort: string) => {
  const direction = sort.startsWith('-') ? -1 : 1
  const value = SORT_VALUES[sort.replace('-', '')] ?? SORT_VALUES.name
  return (a: TenantListItem, b: TenantListItem) => {
    const left = value(a)
    const right = value(b)
    if (left === right) return 0
    if (left === null) return 1
    if (right === null) return -1
    return left < right ? -direction : direction
  }
}

export const tenantHandlers = [
  http.get(`${API}/tenants`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const q = params.get('q')?.toLowerCase()
    const categoryId = params.get('category_id')
    const floor = params.get('floor')
    const isActive = params.get('is_active')
    const filtered = tenants
      .filter(
        (t) =>
          (!q || `${t.code} ${t.name}`.toLowerCase().includes(q)) &&
          (!categoryId || t.categoryId === categoryId) &&
          (floor === null || String(t.floor) === floor) &&
          (isActive === null || String(t.isActive) === isActive),
      )
      .sort(compareBySort(params.get('sort') ?? 'name'))
    return HttpResponse.json(paged(request, filtered))
  }),

  http.get(`${API}/tenants/:id`, ({ params }) => {
    const tenant = tenants.find((t) => t.id === params.id)
    if (!tenant) return apiError(404, 'not_found', 'Арендатор не найден')
    return HttpResponse.json(tenant)
  }),

  http.post(`${API}/tenants`, async ({ request }) => {
    const body = (await request.json()) as TenantCreateRequest
    if (isCodeTaken(body.code)) {
      return apiError(409, 'conflict', 'Конфликт состояния')
    }
    const now = new Date().toISOString()
    const tenant: TenantListItem = {
      ...body,
      id: mockId('tenant-new'),
      categoryName: 'Новая категория',
      priority: body.priority ?? 50,
      isActive: body.isActive ?? true,
      createdAt: now,
      updatedAt: now,
      guestsCount: null,
      visitsCount: null,
      newGuestsCount: null,
      suppressed: true,
      sourcesCount: 0,
      problemSourcesCount: 0,
    }
    tenants.push(tenant)
    return HttpResponse.json(tenant, { status: 201 })
  }),

  http.patch(`${API}/tenants/:id`, async ({ params, request }) => {
    const tenant = tenants.find((t) => t.id === params.id)
    if (!tenant) return apiError(404, 'not_found', 'Арендатор не найден')
    const body = (await request.json()) as TenantUpdateRequest
    Object.assign(tenant, body, { updatedAt: new Date().toISOString() })
    if (body.isActive === false) tenant.sourcesCount = 0
    return HttpResponse.json(tenant)
  }),

  http.delete(`${API}/tenants/:id`, ({ params }) => {
    const index = tenants.findIndex((t) => t.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'Арендатор не найден')
    if (params.id === TENANT_WITH_EVENTS_ID) {
      return apiError(409, 'tenant_has_events', 'У арендатора есть события, удаление невозможно')
    }
    tenants.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
