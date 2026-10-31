import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { server } from '@/app/mocks/server'
import { GuestProfilePage } from './guest-profile-page'

function renderPage(id: string, search = '') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[`/guests/${id}${search}`]}>
        <Routes>
          <Route path="/guests/:id" element={<GuestProfilePage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const SESSION_KEY = 'ucrp-mock-session'

describe('GuestProfilePage', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, mockUsers[0].userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('shows the key numbers and the overview tab by default', async () => {
    renderPage('guest-0001')
    expect(await screen.findByRole('heading', { name: 'Иван Петров' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Гости' })).toHaveAttribute('href', '/guests')
    expect(screen.getByText('Сумма покупок')).toBeInTheDocument()
    expect(await screen.findByText('Следующий шаг')).toBeInTheDocument()
    expect(screen.getByText('Постоянные гости')).toBeInTheDocument()
  })

  it('shows the movement map as a chronological chain of floors', async () => {
    renderPage('guest-0001')
    const route = await screen.findByRole('list', { name: 'Маршрут по этажам' })
    const legs = within(route).getAllByRole('listitem')
    expect(legs).toHaveLength(4)
    expect(within(legs[0]).getByText('2 этаж')).toBeInTheDocument()
    expect(within(legs[1]).getByText('4 этаж')).toBeInTheDocument()
    expect(within(legs[1]).getByText('Киномир')).toBeInTheDocument()
    expect(within(legs[2]).getByText('1 этаж')).toBeInTheDocument()
    expect(within(legs[3]).getByText('Северный ТЦ')).toBeInTheDocument()
  })

  it('shows an empty movement map for a guest without movement', async () => {
    renderPage('guest-0002')
    expect(await screen.findByText('Передвижений пока нет')).toBeInTheDocument()
  })

  it('shows recommendations with an explanation popover', async () => {
    renderPage('guest-0001')
    expect(await screen.findByText(/Кинотеатр «Синема»/)).toBeInTheDocument()
    await userEvent.click((await screen.findAllByRole('button', { name: 'Почему?' }))[0])
    expect(await screen.findByText(/Похожие гости посещают кинотеатр/)).toBeInTheDocument()
    expect(screen.getByText('Смежные категории')).toBeInTheDocument()
  })

  it('shows an empty state when there are no recommendations', async () => {
    renderPage('guest-0002')
    expect(await screen.findByText('Рекомендаций пока нет')).toBeInTheDocument()
  })

  it('lists events on the history tab and filters them by type', async () => {
    renderPage('guest-0001')
    await userEvent.click(await screen.findByRole('tab', { name: 'История' }))
    const table = await screen.findByRole('table')
    expect(await within(table).findByText('Лента')).toBeInTheDocument()
    expect(within(table).getByText('Зара')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('combobox', { name: 'Тип события' }))
    await userEvent.click(await screen.findByRole('option', { name: 'Покупка' }))
    await waitFor(() => expect(within(table).queryByText('Зара')).not.toBeInTheDocument())
    expect(within(table).getByText('Лента')).toBeInTheDocument()
  })

  it('opens the purchases tab from the url and shows only purchases', async () => {
    renderPage('guest-0001', '?tab=purchases')
    const table = await screen.findByRole('table')
    expect(await within(table).findByText('Лента')).toBeInTheDocument()
    expect(within(table).queryByText('Зара')).not.toBeInTheDocument()
    expect(within(table).queryByRole('columnheader', { name: 'Событие' })).not.toBeInTheDocument()
  })

  it('shows tenant totals on the tenants tab', async () => {
    renderPage('guest-0001', '?tab=tenants')
    expect(await screen.findByText('Продукты')).toBeInTheDocument()
    expect(await screen.findByRole('columnheader', { name: 'Потрачено' })).toBeInTheDocument()
  })

  it('shows identifiers and contact history on the profile tab', async () => {
    renderPage('guest-0001', '?tab=profile')
    expect(await screen.findByText('79001234567')).toBeInTheDocument()
    expect(await screen.findByText('Email', { selector: 'td' })).toBeInTheDocument()
  })

  it('hides the history content when the field is forbidden', async () => {
    server.use(
      http.get(`${API}/guests/:id/events`, () =>
        apiError(403, 'forbidden_field', 'Нет доступа к полю'),
      ),
    )
    renderPage('guest-0001', '?tab=history')
    expect(await screen.findByRole('heading', { name: 'Иван Петров' })).toBeInTheDocument()
    await waitFor(() => expect(screen.queryByRole('table')).not.toBeInTheDocument())
  })

  it('hides recommendations and history from roles without access', async () => {
    localStorage.setItem(SESSION_KEY, mockUsers[1].userId)
    renderPage('guest-0001')
    expect(await screen.findByRole('heading', { name: 'Иван Петров' })).toBeInTheDocument()
    expect(screen.queryByText('Следующий шаг')).not.toBeInTheDocument()
  })

  it('shows an error for an unknown guest', async () => {
    renderPage('missing')
    expect(await screen.findByText('Гость не найден')).toBeInTheDocument()
  })

  it('offers deletion only with the delete permission and asks for confirmation', async () => {
    renderPage('guest-0001')
    expect(await screen.findByRole('heading', { name: 'Иван Петров' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Удалить' })).not.toBeInTheDocument()
  })

  it('confirms before deleting a guest', async () => {
    localStorage.setItem(SESSION_KEY, mockUsers[3].userId)
    renderPage('guest-0001')
    await userEvent.click(await screen.findByRole('button', { name: 'Удалить' }))
    expect(await screen.findByText('Удалить профиль гостя')).toBeInTheDocument()
  })
})
