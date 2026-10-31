import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import type { TenantVisitAggregate } from '../model/types'

export const OWN_TENANT_ID = 'tenant-0001'

export const mockAggregate: TenantVisitAggregate = { visitsCount: null, suppressed: true }

export const tenantAggregatesHandlers = [
  http.get(`${API}/tenants/:id/aggregates`, ({ params }) => {
    if (params.id !== OWN_TENANT_ID) return apiError(403, 'forbidden_tenant', 'Чужой арендатор')
    return HttpResponse.json(mockAggregate)
  }),
]
