import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { CampaignsPage } from './campaigns-page'

function renderPage(url = '/campaigns') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/campaigns" element={<CampaignsPage />} />
          <Route path="/campaigns/:id" element={<div>Карточка кампании</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('CampaignsPage', () => {
  it('restores the status filter from the URL', async () => {
    renderPage('/campaigns?status=draft')
    expect(await screen.findByText('Удержание клиентов')).toBeInTheDocument()
    expect(screen.queryByText('Высокий чек в «Заре»')).not.toBeInTheDocument()
  })

  it('requests the page, page size and status from the URL', async () => {
    const requests: URLSearchParams[] = []
    server.use(
      http.get('*/api/v1/campaigns', ({ request }) => {
        requests.push(new URL(request.url).searchParams)
        return HttpResponse.json({
          items: [],
          meta: { page: 2, pageSize: 20, total: 45, totalPages: 3 },
        })
      }),
    )
    renderPage('/campaigns?page=2&page_size=20&status=draft')
    await waitFor(() => expect(requests.length).toBeGreaterThan(0))
    const last = requests[requests.length - 1]
    expect(last.get('page')).toBe('2')
    expect(last.get('page_size')).toBe('20')
    expect(last.get('status')).toBe('draft')
  })

  it('opens the campaign on row click', async () => {
    renderPage()
    await userEvent.click(await screen.findByText('Высокий чек в «Заре»'))
    expect(await screen.findByText('Карточка кампании')).toBeInTheDocument()
  })
})
