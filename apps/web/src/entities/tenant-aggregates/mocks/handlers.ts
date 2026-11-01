import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import type { TenantAggregateDay, TenantVisitAggregate } from '../model/types'

export const OWN_TENANT_ID = 'tenant-0001'
export const DATA_TENANT_ID = 'tenant-0002'

const SUPPRESSION_THRESHOLD = 5

export const mockAggregate: TenantVisitAggregate = {
  visitsCount: null,
  guestsCount: null,
  newGuestsCount: null,
  repeatVisitsCount: null,
  revenueTotalMinor: null,
  avgCheckMinor: null,
  firstPurchasesCount: null,
  suppressed: true,
  suppressionThreshold: SUPPRESSION_THRESHOLD,
  series: [],
}

const DAY_MS = 86_400_000

function buildSeries(from: string, to: string): TenantAggregateDay[] {
  const days: TenantAggregateDay[] = []
  const end = Date.parse(to)
  for (let time = Date.parse(from), index = 0; time <= end; time += DAY_MS, index += 1) {
    days.push({
      date: new Date(time).toISOString().slice(0, 10),
      visitsCount: index % 5 === 0 ? null : 10 + index,
      newGuestsCount: index % 3 === 0 ? null : 5 + (index % 4),
      repeatVisitsCount: index % 5 === 0 ? null : 5 + (index % 6),
      revenueTotalMinor: index % 4 === 0 ? null : 150_000 + index * 2_500,
    })
  }
  return days
}

function buildAggregate(from: string | null, to: string | null): TenantVisitAggregate {
  const series = from && to ? buildSeries(from, to) : []
  return {
    from: from ?? undefined,
    to: to ?? undefined,
    visitsCount: 130,
    guestsCount: 48,
    newGuestsCount: 21,
    repeatVisitsCount: 82,
    revenueTotalMinor: 4_512_000,
    avgCheckMinor: 96_000,
    firstPurchasesCount: 17,
    suppressed: true,
    suppressionThreshold: SUPPRESSION_THRESHOLD,
    series,
  }
}

export const tenantAggregatesHandlers = [
  http.get(`${API}/tenants/:id/aggregates`, ({ params, request }) => {
    const search = new URL(request.url).searchParams
    if (params.id === OWN_TENANT_ID) return HttpResponse.json(mockAggregate)
    if (params.id !== DATA_TENANT_ID) return apiError(403, 'forbidden_tenant', 'Чужой арендатор')
    return HttpResponse.json(buildAggregate(search.get('from'), search.get('to')))
  }),
]
