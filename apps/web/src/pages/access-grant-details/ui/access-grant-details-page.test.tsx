import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { userKeys, type Me } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { AccessGrantDetailsPage } from './access-grant-details-page'

const manager = mockUsers[1]
const reader: Me = { ...manager, permissions: ['access.grant.read'] }

function renderPage(url: string, me: Me = manager) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  client.setQueryData(userKeys.me, me)
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/access/grants/:id" element={<AccessGrantDetailsPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('AccessGrantDetailsPage', () => {
  it('shows parameters, permissions and scopes of the grant', async () => {
    renderPage('/access/grants/grant-0002')
    expect(
      await screen.findByRole('heading', { name: 'Грант: Бистро «Уголок»' }),
    ).toBeInTheDocument()
    expect(screen.getByText('События своих гостей')).toBeInTheDocument()
    expect(await screen.findByText('Лента событий')).toBeInTheDocument()
    expect(screen.getByText('Глубина периода, дней: 90')).toBeInTheDocument()
    expect(screen.getByText('Арендатор: Бистро «Уголок»')).toBeInTheDocument()
  })

  it('offers only the transitions allowed for the status', async () => {
    renderPage('/access/grants/grant-0001')
    expect(await screen.findByRole('button', { name: 'Приостановить' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Отозвать' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Изменить' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Активировать' })).not.toBeInTheDocument()
  })

  it('hides every action from a read-only holder', async () => {
    renderPage('/access/grants/grant-0001', reader)
    await screen.findByRole('heading', { name: 'Грант: Заря' })
    expect(screen.queryByRole('button', { name: 'Отозвать' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Изменить' })).not.toBeInTheDocument()
  })

  it('shows the grant history tab from the URL', async () => {
    renderPage('/access/grants/grant-0001?tab=history')
    expect(await screen.findByText('Грант создан')).toBeInTheDocument()
    expect(
      screen.getByText('Действует · Причина: Пилот: агрегаты для арендатора'),
    ).toBeInTheDocument()
  })

  it('asks for a reason before revoking and freezes the revoked grant', async () => {
    renderPage('/access/grants/grant-0002')
    await userEvent.click(await screen.findByRole('button', { name: 'Отозвать' }))
    const dialog = screen.getByRole('dialog')
    expect(within(dialog).getByText(/Отзыв необратим/)).toBeInTheDocument()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Отозвать' }))
    expect(await within(dialog).findByText('Укажите причину')).toBeInTheDocument()
    await userEvent.type(within(dialog).getByLabelText('Причина'), 'Договор не подписан')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Отозвать' }))
    expect(await screen.findByText('Отозван')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Отозвать' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Изменить' })).not.toBeInTheDocument()
  })

  it('reports a missing grant', async () => {
    renderPage('/access/grants/grant-missing')
    expect((await screen.findAllByText('Не найдено')).length).toBeGreaterThan(0)
  })
})
