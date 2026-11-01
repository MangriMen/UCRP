import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { AccessGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { ImportsPage } from './imports-page'

function renderAs(userId: string, url: string) {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/imports"
            element={
              <AccessGuard>
                <ImportsPage />
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

describe('ImportsPage', () => {
  it('requests the page, page size and status from the URL', async () => {
    const requests: URLSearchParams[] = []
    server.use(
      http.get('*/api/v1/imports', ({ request }) => {
        requests.push(new URL(request.url).searchParams)
        return HttpResponse.json({
          items: [],
          meta: { page: 2, pageSize: 20, total: 45, totalPages: 3 },
        })
      }),
    )
    renderAs('user-marketer', '/imports?page=2&page_size=20&status=failed')
    await waitFor(() => expect(requests.length).toBeGreaterThan(0))
    const last = requests[requests.length - 1]
    expect(last.get('page')).toBe('2')
    expect(last.get('page_size')).toBe('20')
    expect(last.get('status')).toBe('failed')
  })
})
