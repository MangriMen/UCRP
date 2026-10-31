import { http, HttpResponse } from 'msw'
import { API, paged } from '@/shared/api/mocks/helpers'
import { mockAuditLogs } from './fixtures'

export const auditLogHandlers = [
  http.get(`${API}/audit-logs`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const entityType = params.get('entity_type')
    const entityId = params.get('entity_id')
    const actor = params.get('actor')
    const action = params.get('action')
    const from = params.get('from')
    const to = params.get('to')
    const filtered = mockAuditLogs.filter(
      (log) =>
        (!entityType || log.entityType === entityType) &&
        (!entityId || log.entityId === entityId) &&
        (!actor || log.actorUserId === actor) &&
        (!action || log.action === action) &&
        (!from || log.createdAt >= from) &&
        (!to || log.createdAt <= to),
    )
    return HttpResponse.json(paged(request, filtered))
  }),
]
