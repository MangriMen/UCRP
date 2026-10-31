import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { mockDataSources } from '@/entities/data-source/mocks'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { API } from '@/shared/api/mocks/helpers'
import { TenantDetailsPage } from './tenant-details-page'

const SESSION_KEY = 'ucrp-mock-session'
const [marketer, , , analyst] = mockUsers

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderPage(url: string) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/tenants/:id"
            element={
              <>
                <TenantDetailsPage />
                <Search />
              </>
            }
          />
          <Route path="/tenants" element={<div>Список арендаторов</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('TenantDetailsPage', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, marketer.userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('shows statistics and the daily chart for the requested period', async () => {
    renderPage('/tenants/tenant-0002?from=2026-10-01&to=2026-10-07')
    expect(await screen.findByTestId('visits-count')).toHaveTextContent('130')
    expect(screen.getByTestId('guests-count')).toHaveTextContent('48')
    expect(screen.getByTestId('new-guests-count')).toHaveTextContent('21')
    expect(screen.getByRole('img', { name: /Динамика по дням: 7 дн\./ })).toBeInTheDocument()
  })

  it('explains suppressed values', async () => {
    renderPage('/tenants/tenant-0001')
    expect(await screen.findByText('Часть значений скрыта')).toBeInTheDocument()
    expect(screen.getByTestId('visits-count')).toHaveTextContent('меньше 5')
  })

  it('rejects an inverted period without querying', async () => {
    renderPage('/tenants/tenant-0002?from=2026-10-07&to=2026-10-01')
    expect(await screen.findByText('Начало периода позже конца')).toBeInTheDocument()
    expect(screen.queryByTestId('visits-count')).not.toBeInTheDocument()
  })

  it('keeps the tab in the URL and lists the guests of the tenant', async () => {
    renderPage('/tenants/tenant-0002')
    await screen.findByRole('heading', { name: 'Бистро «Уголок»' })
    await userEvent.click(await screen.findByRole('tab', { name: 'Гости' }))
    expect(screen.getByTestId('search')).toHaveTextContent('tab=guests')
    expect(await screen.findByText('Без имени')).toBeInTheDocument()
    expect(screen.queryByText('Мария Иванова')).not.toBeInTheDocument()
  })

  it('shows the sources of the tenant with the latest import status', async () => {
    renderPage('/tenants/tenant-0001?tab=sources')
    expect(await screen.findByRole('link', { name: 'Все источники данных' })).toBeInTheDocument()
    expect(await screen.findAllByRole('row')).not.toHaveLength(0)
  })

  it('pages the sources on the server and keeps the tab in the URL', async () => {
    const requestedPages: (string | null)[] = []
    server.use(
      http.get(`${API}/data-sources`, ({ request }) => {
        const params = new URL(request.url).searchParams
        requestedPages.push(params.get('page'))
        return HttpResponse.json({
          items: mockDataSources.slice(0, 1),
          meta: { page: Number(params.get('page') ?? 1), pageSize: 50, total: 120, totalPages: 3 },
        })
      }),
    )
    renderPage('/tenants/tenant-0001?tab=sources')
    await userEvent.click(await screen.findByRole('button', { name: 'Вперёд' }))
    await waitFor(() => expect(requestedPages).toContain('2'))
    expect(screen.getByTestId('search')).toHaveTextContent('tab=sources')
    expect(screen.getByTestId('search')).toHaveTextContent('page=2')
  })

  it('hides editing from a role without reference.write', async () => {
    renderPage('/tenants/tenant-0002?tab=settings')
    expect(await screen.findByText('Питание')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Удалить' })).not.toBeInTheDocument()
  })
})

describe('TenantDetailsPage settings', () => {
  beforeEach(() => {
    localStorage.setItem(SESSION_KEY, analyst.userId)
  })

  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('shows the conflict when deleting a tenant with events', async () => {
    renderPage('/tenants/tenant-0001?tab=settings')
    await userEvent.click(await screen.findByRole('button', { name: 'Удалить' }))
    await userEvent.click(
      within(screen.getByRole('alertdialog')).getByRole('button', { name: 'Удалить' }),
    )
    expect(
      await screen.findByText('У арендатора есть события, удаление невозможно'),
    ).toBeInTheDocument()
  })

  it('deactivates a tenant after confirmation', async () => {
    renderPage('/tenants/tenant-0002?tab=settings')
    await userEvent.click(await screen.findByRole('button', { name: 'Деактивировать' }))
    await userEvent.click(
      within(screen.getByRole('alertdialog')).getByRole('button', { name: 'Деактивировать' }),
    )
    expect(await screen.findByRole('button', { name: 'Активировать' })).toBeInTheDocument()
  })
})
