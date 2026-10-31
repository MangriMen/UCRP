import { http, HttpResponse } from 'msw'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type { Tenant, TenantCreateRequest, TenantUpdateRequest } from '../model/types'
import { mockTenants, TENANT_WITH_EVENTS_ID } from './fixtures'

const tenants: Tenant[] = structuredClone(mockTenants)

const isCodeTaken = (code: string, exceptId?: string) =>
  tenants.some((t) => t.code === code && t.id !== exceptId)

export const tenantHandlers = [
  http.get(`${API}/tenants`, ({ request }) => {
    const q = new URL(request.url).searchParams.get('q')?.toLowerCase()
    const filtered = q
      ? tenants.filter((t) => `${t.code} ${t.name}`.toLowerCase().includes(q))
      : tenants
    return HttpResponse.json(paged(request, filtered))
  }),

  http.post(`${API}/tenants`, async ({ request }) => {
    const body = (await request.json()) as TenantCreateRequest
    if (isCodeTaken(body.code)) {
      return apiError(409, 'duplicate', 'Арендатор с таким кодом уже существует')
    }
    const now = new Date().toISOString()
    const tenant: Tenant = {
      ...body,
      id: mockId('tenant-new'),
      categoryName: 'Новая категория',
      priority: body.priority ?? 50,
      isActive: body.isActive ?? true,
      createdAt: now,
      updatedAt: now,
    }
    tenants.push(tenant)
    return HttpResponse.json(tenant, { status: 201 })
  }),

  http.patch(`${API}/tenants/:id`, async ({ params, request }) => {
    const tenant = tenants.find((t) => t.id === params.id)
    if (!tenant) return apiError(404, 'not_found', 'Арендатор не найден')
    const body = (await request.json()) as TenantUpdateRequest
    if (body.code && isCodeTaken(body.code, tenant.id)) {
      return apiError(409, 'duplicate', 'Арендатор с таким кодом уже существует')
    }
    Object.assign(tenant, body, { updatedAt: new Date().toISOString() })
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
