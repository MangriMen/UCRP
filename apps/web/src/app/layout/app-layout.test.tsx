import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { setAccessToken } from '@/shared/api/client'
import { AppLayout } from './app-layout'
import { HomeRedirect } from './home-redirect'

const SESSION_KEY = 'ucrp-mock-session'

function renderApp(userId: string) {
  localStorage.setItem(SESSION_KEY, userId)
  const router = createMemoryRouter(
    [
      {
        element: <AppLayout />,
        children: [
          { index: true, element: <HomeRedirect /> },
          { path: '/tenant', element: <p>tenant page</p> },
          { path: '/dashboard', element: <p>dashboard page</p> },
        ],
      },
    ],
    { initialEntries: ['/'] },
  )
  render(
    <QueryClientProvider client={new QueryClient()}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

afterEach(() => {
  localStorage.clear()
  setAccessToken(null)
})

describe('AppLayout', () => {
  it('sends the tenant to their only section and shows one menu entry', async () => {
    renderApp('user-tenant')
    expect(await screen.findByText('tenant page')).toBeInTheDocument()
    expect(screen.getAllByRole('link')).toHaveLength(1)
  })

  it('sends the manager to the dashboard', async () => {
    renderApp('user-manager')
    expect(await screen.findByText('dashboard page')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Пользователи' })).toBeInTheDocument()
  })
})
