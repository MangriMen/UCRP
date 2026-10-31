import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { AccessGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { ExportsPage } from './exports-page'

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderAs(userId: string, url = '/exports') {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/exports"
            element={
              <AccessGuard>
                <ExportsPage />
                <Search />
              </AccessGuard>
            }
          />
          <Route path="/exports/:id" element={<div>Карточка выгрузки</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

afterEach(() => {
  localStorage.clear()
  setAccessToken(null)
})

describe('ExportsPage', () => {
  it('offers new export and CSV download to the marketer', async () => {
    renderAs('user-marketer')
    expect((await screen.findAllByRole('button', { name: 'Скачать CSV' })).length).toBeGreaterThan(
      0,
    )
    expect(screen.getByRole('button', { name: 'Выгрузить' })).toBeInTheDocument()
    expect(screen.queryByText('Формат')).not.toBeInTheDocument()
  })

  it.each(['user-analyst', 'user-manager'])('keeps %s read-only', async (userId) => {
    renderAs(userId)
    await waitFor(() => expect(screen.getAllByRole('row').length).toBeGreaterThan(1))
    expect(screen.queryByRole('button', { name: 'Скачать CSV' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Выгрузить' })).not.toBeInTheDocument()
  })

  it('filters the journal by status from the URL', async () => {
    renderAs('user-marketer', '/exports?status=failed')
    await waitFor(() => expect(screen.getAllByRole('row')).toHaveLength(2))
    expect(within(screen.getByRole('table')).getByText('Ошибка')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Сбросить' }))
    expect(screen.getByTestId('search')).toHaveTextContent('')
  })

  it('opens the export card from a row', async () => {
    renderAs('user-manager')
    await userEvent.click((await screen.findAllByText('Спящий сосед · v2'))[0])
    expect(await screen.findByText('Карточка выгрузки')).toBeInTheDocument()
  })
})
