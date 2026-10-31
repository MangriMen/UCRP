import type { Schema } from './types'

const BASE_URL = '/api/v1'
const REQUEST_TIMEOUT_MS = 30_000

export type ApiErrorBody = Omit<Schema<'ErrorResponse'>['error'], 'request_id'> & {
  request_id?: string
}
export type ApiErrorCode = ApiErrorBody['code']

export class ApiError extends Error {
  readonly status: number
  readonly body: ApiErrorBody

  constructor(status: number, body: ApiErrorBody) {
    super(body.message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }
}

let accessToken: string | null = null
let onUnauthorized: (() => void) | null = null
let refreshing: Promise<boolean> | null = null

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function setUnauthorizedHandler(handler: (() => void) | null) {
  onUnauthorized = handler
}

// Single-flight: параллельные 401 ждут один и тот же запрос refresh.
function refreshAccessToken(): Promise<boolean> {
  refreshing ??= fetch(`${BASE_URL}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
  })
    .then(async (res) => {
      if (!res.ok) return false
      const data = (await res.json()) as Schema<'TokenResponse'>
      accessToken = data.accessToken
      return true
    })
    .catch(() => false)
    .finally(() => {
      refreshing = null
    })
  return refreshing
}

async function toApiError(res: Response): Promise<ApiError> {
  try {
    const data = (await res.json()) as { error?: ApiErrorBody }
    if (data.error) return new ApiError(res.status, data.error)
  } catch {
    // тело не JSON: падаем на общий текст ниже
  }
  return new ApiError(res.status, { code: 'internal_error', message: res.statusText })
}

function toTransportError(e: unknown): unknown {
  if (e instanceof DOMException && e.name === 'TimeoutError') {
    return new ApiError(504, { code: 'timeout', message: 'Сервер не ответил вовремя' })
  }
  return e
}

const NO_REFRESH_PATHS = new Set(['/auth/login', '/auth/refresh'])

export interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown
  query?: Record<string, string | number | boolean | undefined>
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { body, query, headers, ...init } = options
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query ?? {})) {
    if (value !== undefined) params.set(key, String(value))
  }
  const url = `${BASE_URL}${path}${params.size ? `?${params}` : ''}`

  const send = () => {
    const h = new Headers(headers)
    if (accessToken) h.set('Authorization', `Bearer ${accessToken}`)
    if (body !== undefined && !(body instanceof FormData)) h.set('Content-Type', 'application/json')
    const timeout = AbortSignal.timeout(REQUEST_TIMEOUT_MS)
    return fetch(url, {
      ...init,
      signal: init.signal ? AbortSignal.any([init.signal, timeout]) : timeout,
      headers: h,
      credentials: 'include',
      body:
        body === undefined || body instanceof FormData
          ? (body as FormData | undefined)
          : JSON.stringify(body),
    })
  }

  let res: Response
  try {
    res = await send()
  } catch (e) {
    throw toTransportError(e)
  }
  if (res.status === 401 && !NO_REFRESH_PATHS.has(path)) {
    if (await refreshAccessToken()) {
      try {
        res = await send()
      } catch (e) {
        throw toTransportError(e)
      }
    } else {
      onUnauthorized?.()
    }
  }
  if (!res.ok) throw await toApiError(res)
  return res.status === 204 ? (undefined as T) : ((await res.json()) as T)
}
