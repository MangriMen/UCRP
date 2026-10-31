import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import type { Granularity, CrossTenantPoint } from '../model/types'
import { mockMetrics, mockCrossTenantPoints } from './fixtures'

const MAX_PERIOD_MS = 24 * 31 * 24 * 60 * 60 * 1000

function readPeriod(request: Request) {
  const params = new URL(request.url).searchParams
  const from = params.get('from')
  const to = params.get('to')
  if (!from || !to || from > to) return null
  return { from, to, granularity: (params.get('granularity') ?? 'month') as Granularity }
}

function monthPeriods(from: string, to: string) {
  const periods: string[] = []
  const cursor = new Date(from)
  cursor.setUTCDate(1)
  const end = new Date(to)
  while (cursor <= end && periods.length < 12) {
    periods.push(cursor.toISOString().slice(0, 7))
    cursor.setUTCMonth(cursor.getUTCMonth() + 1)
  }
  return periods
}

export const dashboardHandlers = [
  http.get(`${API}/dashboard/metrics`, ({ request }) => {
    const period = readPeriod(request)
    if (!period) return apiError(400, 'validation_error', 'Некорректный период')
    return HttpResponse.json({ from: period.from, to: period.to, ...mockMetrics })
  }),

  http.get(`${API}/dashboard/cross-tenant`, ({ request }) => {
    const period = readPeriod(request)
    if (
      !period ||
      new Date(period.to).getTime() - new Date(period.from).getTime() > MAX_PERIOD_MS
    ) {
      return apiError(400, 'validation_error', 'Период не должен превышать 24 месяца')
    }
    const series: CrossTenantPoint[] = monthPeriods(period.from, period.to).map((name, index) => ({
      period: name,
      ...mockCrossTenantPoints[index % mockCrossTenantPoints.length],
    }))
    const last = series.at(-1) ?? { period: '', ...mockCrossTenantPoints[0] }
    return HttpResponse.json({
      ...period,
      distribution: { one: last.one, two: last.two, threePlus: last.threePlus },
      crossTenantRate: last.crossTenantRate ?? 0,
      series,
    })
  }),
]
