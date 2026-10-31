import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { server } from '@/app/mocks/server'
import { GuestProfilePage } from './guest-profile-page'

function renderPage(id: string) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[`/guests/${id}`]}>
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

  it('shows the card, event feed and contact history', async () => {
    renderPage('guest-0001')
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('79001234567')).toBeInTheDocument()
    expect(await screen.findByText('Покупка')).toBeInTheDocument()
    expect(await screen.findByText('Лента', { selector: 'div' })).toBeInTheDocument()
    expect(await screen.findByText('Email', { selector: 'td' })).toBeInTheDocument()
    expect(await screen.findByText('Следующий шаг')).toBeInTheDocument()
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

  it('hides the event feed when the field is forbidden', async () => {
    server.use(
      http.get(`${API}/guests/:id/events`, () =>
        apiError(403, 'forbidden_field', 'Нет доступа к полю'),
      ),
    )
    renderPage('guest-0001')
    expect(await screen.findByText('История контактов')).toBeInTheDocument()
    await waitFor(() => expect(screen.queryByText('Лента событий')).not.toBeInTheDocument())
  })

  it('hides recommendations from roles without access', async () => {
    localStorage.setItem(SESSION_KEY, mockUsers[1].userId)
    renderPage('guest-0001')
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.queryByText('Следующий шаг')).not.toBeInTheDocument()
  })

  it('shows an error for an unknown guest', async () => {
    renderPage('missing')
    expect(await screen.findByText('Гость не найден')).toBeInTheDocument()
  })
})
