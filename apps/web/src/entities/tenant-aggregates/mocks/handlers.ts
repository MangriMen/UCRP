import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import type { TenantAggregates } from '../model/types'

export const OWN_TENANT_ID = 'tenant-0001'

const SUPPRESSION_THRESHOLD = 5

export const mockAggregates: Omit<TenantAggregates, 'from' | 'to' | 'tenantId'> = {
  tenantName: 'Лента',
  newGuestsViaComplex: 148,
  repeatVisits: null,
  suppressed: true,
  suppressionThreshold: SUPPRESSION_THRESHOLD,
  series: [
    { period: '2026-08', newGuests: 52, repeatVisits: 310 },
    { period: '2026-09', newGuests: null, repeatVisits: 275 },
    { period: '2026-10', newGuests: 61, repeatVisits: null },
  ],
}

export const tenantAggregatesHandlers = [
  http.get(`${API}/tenants/:id/aggregates`, ({ params, request }) => {
    const search = new URL(request.url).searchParams
    const from = search.get('from')
    const to = search.get('to')
    if (!from || !to || from > to) return apiError(400, 'validation_error', 'Некорректный период')
    if (params.id !== OWN_TENANT_ID) return apiError(403, 'forbidden_tenant', 'Чужой арендатор')
    return HttpResponse.json({ tenantId: OWN_TENANT_ID, from, to, ...mockAggregates })
  }),
]
