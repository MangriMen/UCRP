import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import { http } from 'msw'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
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

describe('GuestProfilePage', () => {
  it('shows the card, event feed and contact history', async () => {
    renderPage('guest-0001')
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('79001234567')).toBeInTheDocument()
    expect(await screen.findByText('Покупка')).toBeInTheDocument()
    expect(await screen.findByText('Лента', { selector: 'div' })).toBeInTheDocument()
    expect(await screen.findByText('Email', { selector: 'td' })).toBeInTheDocument()
    expect(screen.getByText('Следующий шаг')).toBeInTheDocument()
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

  it('shows an error for an unknown guest', async () => {
    renderPage('missing')
    expect(await screen.findByText('Гость не найден')).toBeInTheDocument()
  })
})
