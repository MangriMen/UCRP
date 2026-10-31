import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { setAccessToken } from '@/shared/api/client'
import { mockUsers } from '../mocks'
import type { Role } from '../model/types'
import { RoleGuard } from './role-guard'

const SESSION_KEY = 'ucrp-mock-session'

function renderGuard(roles?: Role[]) {
  const queryClient = new QueryClient()
  const router = createMemoryRouter(
    [
      { path: '/login', element: <p>login page</p> },
      {
        path: '/',
        element: (
          <RoleGuard roles={roles}>
            <p>secret</p>
          </RoleGuard>
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

describe('RoleGuard', () => {
  it('redirects to /login without a session', async () => {
    renderGuard()
    expect(await screen.findByText('login page')).toBeInTheDocument()
  })

  it('restores the session through refresh and renders children', async () => {
    localStorage.setItem(SESSION_KEY, mockUsers[0].userId)
    renderGuard(['marketer'])
    expect(await screen.findByText('secret')).toBeInTheDocument()
  })

  it('shows the forbidden state for a disallowed role', async () => {
    localStorage.setItem(SESSION_KEY, mockUsers[3].userId)
    renderGuard(['marketer'])
    expect(await screen.findByText('Недостаточно прав')).toBeInTheDocument()
    expect(screen.queryByText('secret')).not.toBeInTheDocument()
  })
})
