import { http, HttpResponse } from 'msw'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type { Export, ExportCreateRequest, ExportPreview } from '../model/types'
import { mockExports } from './fixtures'

const IN_SEGMENT = 120
const WITH_CONTACT = 100
const BLOCKED_BY_LIMIT = 10

const exports: Export[] = structuredClone(mockExports)

function advance(item: Export) {
  if (item.status === 'building') {
    item.status = 'ready'
    item.exportedCount = item.totalInSegment
    item.finishedAt = new Date().toISOString()
    item.downloadUrl = `/api/v1/exports/${item.id}/download`
  }
}

function preview(controlPercent: number): ExportPreview {
  const control = Math.floor(((WITH_CONTACT - BLOCKED_BY_LIMIT) * controlPercent) / 100)
  return {
    inSegment: IN_SEGMENT,
    withContact: WITH_CONTACT,
    blockedByLimit: BLOCKED_BY_LIMIT,
    control,
    toSend: WITH_CONTACT - BLOCKED_BY_LIMIT - control,
    appliedLimits: [],
  }
}

export const exportHandlers = [
  http.get(`${API}/exports`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const segmentId = params.get('segment_id')
    const channel = params.get('channel')
    const status = params.get('status')
    const filtered = exports.filter(
      (e) =>
        (!segmentId || e.segmentId === segmentId) &&
        (!channel || e.channel === channel) &&
        (!status || e.status === status),
    )
    filtered.forEach(advance)
    return HttpResponse.json(paged(request, filtered))
  }),

  http.get(`${API}/exports/:id`, ({ params }) => {
    const item = exports.find((e) => e.id === params.id)
    if (!item) return apiError(404, 'not_found', 'Выгрузка не найдена')
    advance(item)
    return HttpResponse.json(item)
  }),

  http.get(`${API}/exports/:id/download`, ({ params }) => {
    const item = exports.find((e) => e.id === params.id)
    if (!item) return apiError(404, 'not_found', 'Выгрузка не найдена')
    if (item.status !== 'ready') return apiError(409, 'export_not_ready', 'Выгрузка ещё не готова')
    return new HttpResponse('guest_id,phone\nguest-0001,+70000000001\n', {
      headers: { 'Content-Type': 'text/csv' },
    })
  }),

  http.post(`${API}/exports/preview`, async ({ request }) => {
    const body = (await request.json()) as { controlPercent?: number }
    return HttpResponse.json(preview(body.controlPercent ?? 0))
  }),

  http.post(`${API}/exports`, async ({ request }) => {
    if (!request.headers.get('Idempotency-Key')) {
      return apiError(400, 'validation_error', 'Нужен заголовок Idempotency-Key')
    }
    const body = (await request.json()) as ExportCreateRequest
    const counts = preview(body.controlPercent ?? 0)
    const item: Export = {
      id: mockId('export-new'),
      segmentId: body.segmentId,
      segmentName: body.segmentId,
      segmentVersionId: `${body.segmentId}-v1`,
      segmentVersionNo: 1,
      campaignId: body.campaignId ?? null,
      channel: body.channel,
      format: body.format,
      status: 'building',
      controlPercent: body.controlPercent ?? 0,
      totalInSegment: counts.inSegment,
      withContactCount: counts.withContact,
      noContactCount: counts.inSegment - counts.withContact,
      blockedByLimitCount: counts.blockedByLimit,
      controlCount: counts.control,
      exportedCount: 0,
      createdAt: new Date().toISOString(),
      finishedAt: null,
      downloadUrl: null,
    }
    exports.unshift(item)
    return HttpResponse.json(item, { status: 201 })
  }),
]
