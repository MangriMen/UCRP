import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { AccessGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { SegmentsPage } from './segments-page'

function renderAs(userId: string, url = '/segments') {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/segments"
            element={
              <AccessGuard>
                <SegmentsPage />
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

describe('SegmentsPage', () => {
  it.each([
    ['user-marketer', true],
    ['user-analyst', true],
    ['user-manager', false],
  ])('%s can create a segment: %s', async (userId, allowed) => {
    renderAs(userId)
    expect(await screen.findByText('Спящий сосед')).toBeInTheDocument()
    expect(!!screen.queryByRole('button', { name: 'Новый сегмент' })).toBe(allowed)
  })

  it('requests the page, page size and filters from the URL', async () => {
    const requests: URLSearchParams[] = []
    server.use(
      http.get('*/api/v1/segments', ({ request }) => {
        requests.push(new URL(request.url).searchParams)
        return HttpResponse.json({
          items: [],
          meta: { page: 2, pageSize: 20, total: 45, totalPages: 3 },
        })
      }),
    )
    renderAs('user-marketer', '/segments?page=2&page_size=20&q=сосед&status=ready')
    await waitFor(() => expect(requests.length).toBeGreaterThan(0))
    const last = requests[requests.length - 1]
    expect(last.get('page')).toBe('2')
    expect(last.get('page_size')).toBe('20')
    expect(last.get('q')).toBe('сосед')
    expect(last.get('status')).toBe('ready')
  })
})
