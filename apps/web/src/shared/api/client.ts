const BASE_URL = '/api/v1'

export interface ApiErrorBody {
  code: string
  message: string
  details?: { field: string; reason: string }[]
  request_id?: string
}

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
      const data = (await res.json()) as { access_token: string }
      accessToken = data.access_token
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
  return new ApiError(res.status, { code: 'unknown', message: res.statusText })
}

export interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown
  query?: Record<string, string | number | boolean | undefined>
}

export async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { body, query, headers, ...init } = options
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query ?? {})) {
    if (value !== undefined) params.set(key, String(value))
  }
  const url = `${BASE_URL}${path}${params.size ? `?${params}` : ''}`

  const send = () => {
    const h = new Headers(headers)
    if (accessToken) h.set('Authorization', `Bearer ${accessToken}`)
    if (body !== undefined && !(body instanceof FormData))
      h.set('Content-Type', 'application/json')
    return fetch(url, {
      ...init,
      headers: h,
      credentials: 'include',
      body:
        body === undefined || body instanceof FormData
          ? (body as FormData | undefined)
          : JSON.stringify(body),
    })
  }

  let res = await send()
  if (res.status === 401 && !path.startsWith('/auth/')) {
    if (await refreshAccessToken()) {
      res = await send()
    } else {
      onUnauthorized?.()
    }
  }
  if (!res.ok) throw await toApiError(res)
  return res.status === 204 ? (undefined as T) : ((await res.json()) as T)
}
