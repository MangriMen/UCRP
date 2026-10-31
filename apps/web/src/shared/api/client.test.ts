import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, request, requestBlob, setAccessToken } from './client'

const json = (status: number, body?: unknown) =>
  new Response(body === undefined ? null : JSON.stringify(body), { status })

afterEach(() => {
  vi.unstubAllGlobals()
  setAccessToken(null)
})

describe('request', () => {
  it('throws ApiError with the backend error body on 4xx', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(json(403, { error: { code: 'forbidden_role', message: 'нет прав' } })),
    )
    await expect(request('/x')).rejects.toMatchObject({
      status: 403,
      body: { code: 'forbidden_role' },
    })
    await expect(request('/x')).rejects.toBeInstanceOf(ApiError)
  })

  it('refreshes once and retries when two requests get 401 at the same time', async () => {
    setAccessToken('old')
    const fetchMock = vi.fn(async (url: string, init?: RequestInit) => {
      if (url.endsWith('/auth/refresh'))
        return json(200, { accessToken: 'new', tokenType: 'Bearer', expiresIn: 900 })
      const auth = new Headers(init?.headers).get('Authorization')
      return auth === 'Bearer new' ? json(200, { ok: true }) : json(401)
    })
    vi.stubGlobal('fetch', fetchMock)
    await Promise.all([request('/a'), request('/b')])
    expect(fetchMock.mock.calls.filter(([u]) => String(u).endsWith('/auth/refresh'))).toHaveLength(
      1,
    )
  })

  it('treats a refresh response without accessToken as a failed refresh', async () => {
    setAccessToken('old')
    const fetchMock = vi.fn(async (url: string) =>
      url.endsWith('/auth/refresh') ? json(200, { access_token: 'x' }) : json(401),
    )
    vi.stubGlobal('fetch', fetchMock)
    await expect(request('/a')).rejects.toMatchObject({ status: 401 })
  })

  it('returns the response body as a blob with the bearer token', async () => {
    setAccessToken('token')
    const fetchMock = vi.fn().mockResolvedValue(new Response('id,phone\n1,+7'))
    vi.stubGlobal('fetch', fetchMock)
    const blob = await requestBlob('/exports/x/download')
    expect(await blob.text()).toBe('id,phone\n1,+7')
    expect(new Headers(fetchMock.mock.calls[0][1].headers).get('Authorization')).toBe(
      'Bearer token',
    )
  })
})
