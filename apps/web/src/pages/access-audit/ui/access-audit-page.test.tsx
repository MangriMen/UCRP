import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { userKeys } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { AccessAuditPage } from './access-audit-page'

const manager = mockUsers[1]
const analyst = mockUsers.find((user) => user.role === 'analyst')!

function renderPage(url = '/access/audit', me = manager) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  client.setQueryData(userKeys.me, me)
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[url]}>
        <AccessAuditPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('AccessAuditPage', () => {
  it('lists grant changes, simulations and denials with their summary', async () => {
    renderPage()
    expect(await screen.findByText('Отказ в доступе')).toBeInTheDocument()
    expect(
      screen.getByText('Действует → Отозван · Причина: Договор расторгнут'),
    ).toBeInTheDocument()
    expect(screen.getByText('Проверка доступа')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Грант grant-0006' })).toHaveAttribute(
      'href',
      '/access/grants/grant-0006',
    )
  })

  it('restores the actor filter from the URL and resets it', async () => {
    renderPage('/access/audit?actor=user-tenant')
    expect(await screen.findByText('Отказ в доступе')).toBeInTheDocument()
    expect(screen.queryByText('Грант создан')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Сбросить' }))
    expect(await screen.findByText('Грант создан')).toBeInTheDocument()
  })

  it('skips the user directory and the grants link without user and grant read', async () => {
    let usersRequests = 0
    server.use(
      http.get('*/api/v1/users', () => {
        usersRequests += 1
        return HttpResponse.json({ message: 'forbidden' }, { status: 403 })
      }),
    )
    renderPage('/access/audit', analyst)
    expect(await screen.findByText('Отказ в доступе')).toBeInTheDocument()
    expect(usersRequests).toBe(0)
    expect(screen.queryByRole('link', { name: 'Доступы' })).not.toBeInTheDocument()
  })

  it('filters by grant id', async () => {
    renderPage()
    await userEvent.type(await screen.findByLabelText('ID гранта'), 'grant-0006')
    await waitFor(() => expect(screen.queryByText('Отказ в доступе')).not.toBeInTheDocument())
    expect(screen.getByText('Грант отозван')).toBeInTheDocument()
  })
})
