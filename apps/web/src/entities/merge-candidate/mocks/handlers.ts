import { http, HttpResponse } from 'msw'
import { API, apiError, paged } from '@/shared/api/mocks/helpers'
import type { MergeResolution } from '../model/types'
import { mockMergeCandidates, mockProfiles } from './fixtures'

const findCandidate = (id: unknown) => mockMergeCandidates.find((c) => c.id === id)

type Action = 'confirm' | 'reject' | 'revert'

const resolutionByAction: Record<Action, MergeResolution> = {
  confirm: 'confirmed',
  reject: 'rejected',
  revert: 'reverted',
}

const resolvableFrom: Record<Action, MergeResolution[]> = {
  confirm: ['pending'],
  reject: ['pending'],
  revert: ['confirmed', 'auto_merged'],
}

const resolveHandler = (action: Action) =>
  http.post(`${API}/merge-candidates/:id/${action}`, ({ params }) => {
    const candidate = findCandidate(params.id)
    if (!candidate) return apiError(404, 'not_found', 'Кандидат не найден')
    if (!resolvableFrom[action].includes(candidate.resolution)) {
      return apiError(409, 'already_resolved', 'Пара уже обработана')
    }
    return HttpResponse.json({ ...candidate, resolution: resolutionByAction[action] })
  })

export const mergeCandidateHandlers = [
  http.get(`${API}/merge-candidates`, ({ request }) => {
    const resolution = new URL(request.url).searchParams.get('resolution')
    const matched = mockMergeCandidates.filter((c) => !resolution || c.resolution === resolution)
    return HttpResponse.json(paged(request, matched))
  }),

  http.get(`${API}/merge-candidates/:id`, ({ params }) => {
    const candidate = findCandidate(params.id)
    if (!candidate) return apiError(404, 'not_found', 'Кандидат не найден')
    return HttpResponse.json({
      ...candidate,
      guestA: mockProfiles[candidate.guestAId],
      guestB: mockProfiles[candidate.guestBId],
      sharedKeys: [{ idType: 'phone', valueNormalized: '79001234567' }],
      differingFields: ['displayName'],
    })
  }),

  resolveHandler('confirm'),
  resolveHandler('reject'),
  resolveHandler('revert'),
]
