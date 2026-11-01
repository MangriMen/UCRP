import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { setAccessToken } from '@/shared/api/client'
import { mockUsers } from '../mocks'
import type { PermissionCode } from '../model/permissions'
import { AccessGuard } from './access-guard'

const SESSION_KEY = 'ucrp-mock-session'

function renderGuard(permission?: PermissionCode, tenantArea = false) {
  const queryClient = new QueryClient()
  const router = createMemoryRouter(
    [
      { path: '/login', element: <p>login page</p> },
      {
        path: '/',
        element: (
          <AccessGuard permission={permission} tenantArea={tenantArea}>
            <p>secret</p>
          </AccessGuard>
        ),
      },
    ],
    { initialEntries: ['/'] },
  )
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

afterEach(() => {
  localStorage.clear()
  setAccessToken(null)
})

describe('AccessGuard', () => {
  it('redirects to /login without a session', async () => {
    renderGuard()
    expect(await screen.findByText('login page')).toBeInTheDocument()
  })

  it('restores the session through refresh and renders children', async () => {
    localStorage.setItem(SESSION_KEY, mockUsers[0].userId)
    renderGuard('segment.write')
    expect(await screen.findByText('secret')).toBeInTheDocument()
  })

  it('shows the forbidden state without the permission', async () => {
    localStorage.setItem(SESSION_KEY, mockUsers[3].userId)
    renderGuard('user.write')
    expect(await screen.findByText('Недостаточно прав')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'На стартовую страницу' })).toHaveAttribute('href', '/')
    expect(screen.queryByText('secret')).not.toBeInTheDocument()
  })

  it('tells a tenant without a grant to ask the manager', async () => {
    const tenant = mockUsers.find((user) => user.role === 'tenant')
    localStorage.setItem(SESSION_KEY, tenant?.userId ?? '')
    renderGuard('user.write', true)
    expect(await screen.findByText('Доступ ещё не выдан')).toBeInTheDocument()
    expect(screen.queryByText('secret')).not.toBeInTheDocument()
  })
})
