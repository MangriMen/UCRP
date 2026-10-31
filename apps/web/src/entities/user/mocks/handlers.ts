import { http, HttpResponse } from 'msw'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { MOCK_PASSWORD, mockUsers } from './fixtures'

const SESSION_KEY = 'ucrp-mock-session'
const tokenFor = (userId: string) => `mock-token-${userId}`

const sessionUser = () => mockUsers.find((u) => u.userId === localStorage.getItem(SESSION_KEY))

const tokenResponse = (userId: string) => ({
  accessToken: tokenFor(userId),
  tokenType: 'Bearer' as const,
  expiresIn: 900,
})

export const userHandlers = [
  http.post(`${API}/auth/login`, async ({ request }) => {
    const { email, password } = (await request.json()) as { email: string; password: string }
    const user = mockUsers.find((u) => u.email === email)
    if (!user || password !== MOCK_PASSWORD) {
      return apiError(401, 'invalid_credentials', 'Неверный email или пароль')
    }
    localStorage.setItem(SESSION_KEY, user.userId)
    return HttpResponse.json({ ...tokenResponse(user.userId), user })
  }),

  http.post(`${API}/auth/refresh`, () => {
    const user = sessionUser()
    if (!user) return apiError(401, 'unauthorized', 'Сессия истекла')
    return HttpResponse.json(tokenResponse(user.userId))
  }),

  http.post(`${API}/auth/logout`, () => {
    localStorage.removeItem(SESSION_KEY)
    return new HttpResponse(null, { status: 204 })
  }),

  http.get(`${API}/auth/me`, ({ request }) => {
    const user = sessionUser()
    const authorized =
      user && request.headers.get('Authorization') === `Bearer ${tokenFor(user.userId)}`
    if (!authorized) return apiError(401, 'unauthorized', 'Требуется авторизация')
    return HttpResponse.json(user)
  }),
]
