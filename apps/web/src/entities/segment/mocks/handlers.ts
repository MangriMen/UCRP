import { http, HttpResponse } from 'msw'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import { isSegmentComputing } from '../model/status'
import type {
  Segment,
  SegmentCreateRequest,
  SegmentMember,
  SegmentUpdateRequest,
} from '../model/types'
import { mockSegments, mockSegmentTemplates, mockSegmentVersions } from './fixtures'

const READY_MEMBER_COUNT = 150

const segments: Segment[] = structuredClone(mockSegments)

function advance(segment: Segment) {
  if (segment.status === 'pending') {
    segment.status = 'computing'
  } else if (segment.status === 'computing') {
    segment.status = 'ready'
    segment.memberCount = READY_MEMBER_COUNT
    segment.computedAt = new Date().toISOString()
    segment.updatedAt = segment.computedAt
  }
}

const members = (segment: Segment): SegmentMember[] =>
  Array.from({ length: segment.memberCount }, (_, index) => ({
    guestId: `guest-${String(index + 1).padStart(4, '0')}`,
    displayName: `Гость ${index + 1}`,
    tenantCount: (index % 4) + 1,
    lastSeenAt: '2026-10-28T18:20:00Z',
    addedAt: segment.computedAt ?? segment.createdAt,
  }))

export const segmentHandlers = [
  http.get(`${API}/segments`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const status = params.get('status')
    const query = params.get('q')?.trim().toLowerCase()
    const filtered = segments.filter(
      (s) => (!status || s.status === status) && (!query || s.name.toLowerCase().includes(query)),
    )
    filtered.forEach(advance)
    return HttpResponse.json(paged(request, filtered))
  }),

  http.post(`${API}/segments`, async ({ request }) => {
    const body = (await request.json()) as SegmentCreateRequest
    const now = new Date().toISOString()
    const segment: Segment = {
      id: mockId('segment-new'),
      name: body.name,
      description: body.description ?? null,
      templateCode: body.templateCode ?? null,
      isTemplate: false,
      currentVersionNo: 1,
      status: 'pending',
      memberCount: 0,
      computedAt: null,
      rule: body.rule,
      createdAt: now,
      updatedAt: now,
    }
    segments.unshift(segment)
    return HttpResponse.json(segment, { status: 201 })
  }),

  http.get(`${API}/segments/:id`, ({ params }) => {
    const segment = segments.find((s) => s.id === params.id)
    if (!segment) return apiError(404, 'not_found', 'Сегмент не найден')
    advance(segment)
    return HttpResponse.json(segment)
  }),

  http.get(`${API}/segments/:id/members`, ({ params, request }) => {
    const segment = segments.find((s) => s.id === params.id)
    if (!segment) return apiError(404, 'not_found', 'Сегмент не найден')
    return HttpResponse.json(paged(request, members(segment)))
  }),

  http.get(`${API}/segments/:id/versions`, ({ params }) => {
    const segment = segments.find((s) => s.id === params.id)
    if (!segment) return apiError(404, 'not_found', 'Сегмент не найден')
    return HttpResponse.json(mockSegmentVersions(segment))
  }),

  http.put(`${API}/segments/:id`, async ({ params, request }) => {
    const segment = segments.find((s) => s.id === params.id)
    if (!segment) return apiError(404, 'not_found', 'Сегмент не найден')
    const body = (await request.json()) as SegmentUpdateRequest
    segment.name = body.name ?? segment.name
    segment.description = body.description === undefined ? segment.description : body.description
    segment.rule = body.rule
    segment.currentVersionNo += 1
    segment.status = 'pending'
    segment.updatedAt = new Date().toISOString()
    return HttpResponse.json(segment)
  }),

  http.post(`${API}/segments/:id/recompute`, ({ params }) => {
    const segment = segments.find((s) => s.id === params.id)
    if (!segment) return apiError(404, 'not_found', 'Сегмент не найден')
    if (isSegmentComputing(segment.status)) {
      return apiError(409, 'recompute_in_progress', 'Пересчёт уже выполняется')
    }
    segment.status = 'pending'
    return HttpResponse.json(
      {
        jobId: `job-${segment.id}`,
        jobType: 'segment.recompute',
        status: 'queued',
        dedupKey: `segment:${segment.id}:${segment.currentVersionNo}`,
      },
      { status: 202 },
    )
  }),

  http.get(`${API}/segment-templates`, () => HttpResponse.json(mockSegmentTemplates)),
]
