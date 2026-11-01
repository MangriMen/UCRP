import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { AuditLogPage } from './audit-log-page'

function renderPage(url = '/admin/audit') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <AuditLogPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('AuditLogPage', () => {
  it('restores filters from the URL', async () => {
    renderPage('/admin/audit?action=user.update')
    expect(await screen.findByText('user · user-marketer-2')).toBeInTheDocument()
    expect(screen.queryByText('export · export-0001')).not.toBeInTheDocument()
  })

  it('requests the page, page size and filters from the URL', async () => {
    const requests: URLSearchParams[] = []
    server.use(
      http.get('*/api/v1/audit-logs', ({ request }) => {
        requests.push(new URL(request.url).searchParams)
        return HttpResponse.json({
          items: [],
          meta: { page: 2, pageSize: 20, total: 45, totalPages: 3 },
        })
      }),
    )
    renderPage('/admin/audit?page=2&page_size=20&action=user.update')
    await waitFor(() => expect(requests.length).toBeGreaterThan(0))
    const last = requests[requests.length - 1]
    expect(last.get('page')).toBe('2')
    expect(last.get('page_size')).toBe('20')
    expect(last.get('action')).toBe('user.update')
  })

  it('applies the action filter on submit', async () => {
    renderPage()
    expect(await screen.findByText('export · export-0001')).toBeInTheDocument()
    await userEvent.type(screen.getByLabelText('Действие'), 'export.create')
    await userEvent.click(screen.getByRole('button', { name: 'Применить' }))
    expect(await screen.findByText('export · export-0001')).toBeInTheDocument()
    expect(screen.queryByText('user · user-marketer-2')).not.toBeInTheDocument()
  })
})
