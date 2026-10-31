import { HttpResponse } from 'msw'
import type { ApiErrorCode } from '../client'
import type { Schema } from '../types'

export const API = '/api/v1'

export function apiError(
  status: number,
  code: ApiErrorCode,
  message: string,
  details?: Schema<'ErrorResponse'>['error']['details'],
) {
  return HttpResponse.json(
    { error: { code, message, details, request_id: 'mock-request' } },
    { status },
  )
}
export function paged<T>(request: Request, all: T[]) {
  const params = new URL(request.url).searchParams
  const page = Math.max(1, Number(params.get('page') ?? 1))
  const pageSize = Math.min(200, Math.max(1, Number(params.get('page_size') ?? 50)))
  return {
    items: all.slice((page - 1) * pageSize, page * pageSize),
    meta: { page, pageSize, total: all.length, totalPages: Math.ceil(all.length / pageSize) },
  }
}

let seq = 0
export const mockId = (prefix: string) => `${prefix}-${String(++seq).padStart(4, '0')}`
