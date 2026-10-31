import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { DashboardPage } from './dashboard-page'

const SESSION_KEY = 'ucrp-mock-session'
const [marketer] = mockUsers

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderPage(url = '/dashboard') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/dashboard"
            element={
              <>
                <DashboardPage />
                <Search />
              </>
            }
          />
          <Route path="/tenants" element={<div>Список арендаторов</div>} />
          <Route path="/tenants/:id" element={<div>Карточка арендатора</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('DashboardPage', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, marketer.userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('shows metrics with sample sizes and the cross-tenant chart', async () => {
    renderPage()
    expect(await screen.findByTestId('metric-matchRate')).toHaveTextContent('74')
    expect(screen.getByText('событий: 48 200')).toBeInTheDocument()
    expect(screen.getByTestId('metric-segmentToExport')).toHaveTextContent('3 мин 5 с')
    expect((await screen.findAllByRole('img')).length).toBeGreaterThan(0)
  })

  it('does not query the API for an inverted period', async () => {
    renderPage()
    await screen.findByTestId('metric-matchRate')
    fireEvent.change(screen.getByLabelText('С'), { target: { value: '2999-01-01' } })
    expect(await screen.findByText('Укажите корректный период')).toBeInTheDocument()
  })

  it('opens with the period from the URL and writes changes back to it', async () => {
    renderPage('/dashboard?period_from=2026-01-01&period_to=2026-02-01')
    expect(screen.getByLabelText('С')).toHaveValue('2026-01-01')
    expect(screen.getByLabelText('По')).toHaveValue('2026-02-01')
    await screen.findByTestId('metric-matchRate')
    fireEvent.change(screen.getByLabelText('С'), { target: { value: '2026-01-15' } })
    expect(screen.getByTestId('search')).toHaveTextContent(
      '?period_from=2026-01-15&period_to=2026-02-01',
    )
  })

  it('lists the top tenants by visits with suppressed values and links to the card', async () => {
    renderPage('/dashboard?period_from=2026-01-01&period_to=2026-02-01')
    expect(await screen.findByRole('link', { name: 'Киномир' })).toHaveAttribute(
      'href',
      '/tenants/tenant-0003',
    )
    const names = screen
      .getAllByRole('row')
      .slice(1)
      .map((row) => row.textContent ?? '')
    expect(names[0]).toContain('Киномир')
    expect(names[1]).toContain('Бистро «Уголок»')
    expect(names[2]).toContain('Заря')
    const zarya = within(screen.getByRole('link', { name: 'Заря' }).closest('tr') as HTMLElement)
    expect(zarya.getAllByText('меньше 5')).toHaveLength(3)
  })

  it('links to all tenants with the dashboard period', async () => {
    renderPage('/dashboard?period_from=2026-01-01&period_to=2026-02-01')
    expect(await screen.findByRole('link', { name: 'Все арендаторы' })).toHaveAttribute(
      'href',
      '/tenants?period_from=2026-01-01&period_to=2026-02-01',
    )
  })

  it('hides the tenant block without the reference.read permission', async () => {
    localStorage.clear()
    renderPage()
    await screen.findByTestId('metric-matchRate')
    expect(screen.queryByText('Арендаторы за период')).not.toBeInTheDocument()
  })
})
