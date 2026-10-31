import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
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

const SESSION_KEY = 'ucrp-mock-session'
const [marketer, , tenant] = mockUsers

const rowNames = () =>
  screen
    .getAllByRole('row')
    .slice(1)
    .map((row) => row.textContent ?? '')

describe('GuestsPage', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, marketer.userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('lists guests without filters', async () => {
    renderPage()
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('Мария Иванова')).toBeInTheDocument()
  })

  it('filters by tenant and resets', async () => {
    renderPage('/guests?tenant_id=tenant-0002')
    expect(await screen.findByText('Без имени')).toBeInTheDocument()
    expect(screen.queryByText('Мария Иванова')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Сбросить' }))
    expect(await screen.findByText('Мария Иванова')).toBeInTheDocument()
  })

  it('shows matching guests after typing a query', async () => {
    renderPage()
    await userEvent.type(screen.getByLabelText('Поиск гостей'), 'Петров')
    await waitFor(() => expect(screen.queryByText('Мария Иванова')).not.toBeInTheDocument())
    expect(screen.getByText('Иван Петров')).toBeInTheDocument()
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

  it('sorts on the server by the clicked column and keeps it in the URL', async () => {
    renderPage()
    await screen.findByText('Иван Петров')
    await userEvent.click(screen.getByRole('button', { name: 'Событий' }))
    await waitFor(() => expect(rowNames()[0]).toContain('Мария Иванова'))
    expect(rowNames()[2]).toContain('Иван Петров')
    await userEvent.click(screen.getByRole('button', { name: 'Событий' }))
    await waitFor(() => expect(rowNames()[0]).toContain('Иван Петров'))
  })

  it('shows the purchase total only with the amount permission', async () => {
    renderPage()
    expect(await screen.findByRole('button', { name: 'Сумма покупок' })).toBeInTheDocument()
    expect(screen.getByText(/12\s?500/)).toBeInTheDocument()
  })

  it('hides the purchase total without the amount permission', async () => {
    localStorage.setItem(SESSION_KEY, tenant.userId)
    renderPage()
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Сумма покупок' })).not.toBeInTheDocument()
  })

  it('shows the first two masked identifiers and the count of the rest', async () => {
    renderPage()
    expect(await screen.findByRole('columnheader', { name: 'Идентификаторы' })).toBeInTheDocument()
    const row = screen.getByText('Иван Петров').closest('tr')
    expect(row).toHaveTextContent('Телефон: ***4567')
    expect(row).toHaveTextContent('Email: i***@example.com')
    expect(row).not.toHaveTextContent('***1234')
    expect(row).toHaveTextContent('ещё 1')
    expect(screen.queryByText(/79001234567/)).not.toBeInTheDocument()
  })

  it('hides the identifiers column without the identifiers permission', async () => {
    localStorage.setItem(SESSION_KEY, tenant.userId)
    renderPage()
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.queryByRole('columnheader', { name: 'Идентификаторы' })).not.toBeInTheDocument()
    expect(screen.queryByText('***4567')).not.toBeInTheDocument()
  })

  it('keeps only guests with purchases', async () => {
    renderPage('/guests?has_purchases=true')
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.queryByText('Мария Иванова')).not.toBeInTheDocument()
  })

  it('filters by segment', async () => {
    renderPage('/guests?segment_id=segment-0001')
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('Мария Иванова')).toBeInTheDocument()
    expect(screen.queryByText('Без имени')).not.toBeInTheDocument()
  })

  it('applies the inactive preset and clears it on the second click', async () => {
    renderPage()
    await screen.findByText('Иван Петров')
    const preset = screen.getByRole('button', { name: 'Неактивные 90+ дней' })
    await userEvent.click(preset)
    expect(preset).toHaveAttribute('aria-pressed', 'true')
    expect(await screen.findByText('Гости по фильтрам не найдены')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Неактивные 90+ дней' }))
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
  })

  it('offers to reset filters when nothing matches', async () => {
    renderPage('/guests?first_seen_from=2099-01-01')
    expect(await screen.findByText('Гости по фильтрам не найдены')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Сбросить фильтры' }))
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
  })

  it('shows the total and the page size from the URL', async () => {
    renderPage('/guests?page_size=20')
    expect(await screen.findByText('Иван Петров')).toBeInTheDocument()
    expect(screen.getByText('Всего: 3')).toBeInTheDocument()
    expect(screen.getByRole('combobox', { name: 'Размер страницы' })).toHaveTextContent(
      '20 на странице',
    )
  })
})
