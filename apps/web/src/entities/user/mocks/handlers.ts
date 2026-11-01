import { http, HttpResponse } from 'msw'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type {
  InvitationAcceptRequest,
  User,
  UserCreateRequest,
  UserInvitationRequest,
  UserUpdateRequest,
} from '../model/types'
import {
  MOCK_INVITATION_TOKEN,
  MOCK_PASSWORD,
  mockUserRecords,
  mockUsers,
  SELF_USER_ID,
} from './fixtures'

const records: User[] = structuredClone(mockUserRecords)

const usedInvitationTokens = new Set<string>()

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

  http.post(`${API}/auth/invitations/accept`, async ({ request }) => {
    const { token, password } = (await request.json()) as InvitationAcceptRequest
    if (token !== MOCK_INVITATION_TOKEN || usedInvitationTokens.has(token)) {
      return apiError(401, 'unauthorized', 'Приглашение недействительно')
    }
    if (password.length < 8 || password.length > 72) {
      return apiError(400, 'validation_error', 'Пароль должен быть от 8 до 72 символов')
    }
    usedInvitationTokens.add(token)
    return new HttpResponse(null, { status: 204 })
  }),

  http.get(`${API}/auth/me`, ({ request }) => {
    const user = sessionUser()
    const authorized =
      user && request.headers.get('Authorization') === `Bearer ${tokenFor(user.userId)}`
    if (!authorized) return apiError(401, 'unauthorized', 'Требуется авторизация')
    return HttpResponse.json(user)
  }),

  http.get(`${API}/users`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const q = params.get('q')?.toLowerCase()
    const role = params.get('role')
    const isActive = params.get('is_active')
    if (isActive && isActive !== 'true' && isActive !== 'false') {
      return apiError(400, 'validation_error', 'is_active: ожидается true или false')
    }
    const filtered = records.filter(
      (u) =>
        (!role || u.role === role) &&
        (!isActive || u.isActive === (isActive === 'true')) &&
        (!q || `${u.email} ${u.displayName}`.toLowerCase().includes(q)),
    )
    return HttpResponse.json(paged(request, filtered))
  }),

  http.post(`${API}/users`, async ({ request }) => {
    const body = (await request.json()) as UserCreateRequest
    if (records.some((u) => u.email === body.email)) {
      return apiError(409, 'duplicate', 'Пользователь с таким email уже существует')
    }
    if (body.role === 'tenant' && !body.tenantId) {
      return apiError(400, 'validation_error', 'Для роли «Арендатор» нужен арендатор')
    }
    const now = new Date().toISOString()
    const user: User = {
      id: mockId('user-new'),
      email: body.email,
      displayName: body.displayName,
      role: body.role,
      tenantId: body.tenantId ?? null,
      isActive: body.isActive ?? true,
      lastLoginAt: null,
      createdAt: now,
      updatedAt: now,
    }
    records.push(user)
    return HttpResponse.json(user, { status: 201 })
  }),

  http.post(`${API}/users/invitations`, async ({ request }) => {
    const body = (await request.json()) as UserInvitationRequest
    if ((body.role === 'tenant') !== Boolean(body.tenantId)) {
      return apiError(400, 'validation_error', 'Арендатор указывается только для роли «Арендатор»')
    }
    if (records.some((u) => u.email === body.email)) {
      return apiError(409, 'conflict', 'Конфликт состояния')
    }
    return new HttpResponse(null, { status: 202 })
  }),

  http.post(`${API}/users/:id/recovery`, ({ params }) => {
    const user = records.find((u) => u.id === params.id)
    if (!user) return apiError(404, 'not_found', 'Пользователь не найден')
    if (!user.isActive) return apiError(403, 'forbidden_role', 'Нет прав')
    return new HttpResponse(null, { status: 202 })
  }),

  http.patch(`${API}/users/:id`, async ({ params, request }) => {
    const user = records.find((u) => u.id === params.id)
    if (!user) return apiError(404, 'not_found', 'Пользователь не найден')
    const body = (await request.json()) as UserUpdateRequest
    if (user.id === SELF_USER_ID && body.isActive === false) {
      return apiError(409, 'conflict', 'Нельзя деактивировать самого себя')
    }
    Object.assign(user, {
      displayName: body.displayName ?? user.displayName,
      role: body.role ?? user.role,
      tenantId: body.tenantId === undefined ? user.tenantId : body.tenantId,
      isActive: body.isActive ?? user.isActive,
      updatedAt: new Date().toISOString(),
    })
    return HttpResponse.json(user)
  }),

  http.delete(`${API}/users/:id`, ({ params }) => {
    const user = records.find((u) => u.id === params.id)
    if (!user) return apiError(404, 'not_found', 'Пользователь не найден')
    if (user.id === SELF_USER_ID) {
      return apiError(409, 'conflict', 'Нельзя деактивировать самого себя')
    }
    user.isActive = false
    return new HttpResponse(null, { status: 204 })
  }),
]
