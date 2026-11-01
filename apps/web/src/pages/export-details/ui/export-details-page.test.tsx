import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { AccessGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { ExportDetailsPage } from './export-details-page'

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderAs(userId: string, url: string) {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/exports/:id"
            element={
              <AccessGuard>
                <ExportDetailsPage />
                <Search />
              </AccessGuard>
            }
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

afterEach(() => {
  localStorage.clear()
  setAccessToken(null)
})

describe('ExportDetailsPage', () => {
  it('shows counters and the recipients page from the URL', async () => {
    renderAs('user-marketer', '/exports/export-0001?page=2&page_size=20')
    expect(
      await screen.findByRole('heading', { name: 'Выгрузка «Спящий сосед»' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Исключено лимитом')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Скачать CSV' })).toBeInTheDocument()
    const table = await screen.findByRole('table')
    expect(await within(table).findByText('Гость 0021')).toBeInTheDocument()
    expect(within(table).getByText('79000000021')).toBeInTheDocument()
  })

  it('keeps the control group tab in the URL', async () => {
    renderAs('user-marketer', '/exports/export-0001')
    await userEvent.click(await screen.findByRole('tab', { name: 'Контрольная группа' }))
    expect(screen.getByTestId('search')).toHaveTextContent('tab=control')
    const table = await screen.findByRole('table')
    expect(await within(table).findByText('Гость 0082')).toBeInTheDocument()
    expect(within(table).queryByText('Гость 0001')).not.toBeInTheDocument()
  })

  it('hides members and download from the manager', async () => {
    renderAs('user-manager', '/exports/export-0001')
    expect(
      await screen.findByRole('heading', { name: 'Выгрузка «Спящий сосед»' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('tab', { name: 'Получатели' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Скачать CSV' })).not.toBeInTheDocument()
  })

  it('shows not found for an unknown export', async () => {
    renderAs('user-marketer', '/exports/export-9999')
    expect(await screen.findByText(/Выгрузка не найдена/)).toBeInTheDocument()
  })
})
