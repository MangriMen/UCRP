import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { GuestsPage } from './guests-page'

function renderPage(url = '/guests') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/guests" element={<GuestsPage />} />
          <Route path="/guests/:id" element={<div>Профиль гостя</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('GuestsPage', () => {
  it('asks for a query before searching', () => {
    renderPage()
    expect(screen.getByText('Введите запрос для поиска')).toBeInTheDocument()
  })

  it('shows matching guests after typing a query', async () => {
    renderPage()
    await userEvent.type(screen.getByLabelText('Поиск гостей'), 'Петров')
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.queryByText('Мария Иванова')).not.toBeInTheDocument()
  })

  it('restores the search from the URL', async () => {
    renderPage('/guests?q=Мария')
    expect(await screen.findByText('Мария Иванова')).toBeInTheDocument()
    expect(screen.getByLabelText('Поиск гостей')).toHaveValue('Мария')
  })

  it('opens the guest profile on row click', async () => {
    renderPage('/guests?q=Мария')
    await userEvent.click(await screen.findByText('Мария Иванова'))
    expect(await screen.findByText('Профиль гостя')).toBeInTheDocument()
  })
})
