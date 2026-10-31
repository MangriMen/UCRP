import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
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

  it('applies the action filter on submit', async () => {
    renderPage()
    expect(await screen.findByText('export · export-0001')).toBeInTheDocument()
    await userEvent.type(screen.getByLabelText('Действие'), 'export.create')
    await userEvent.click(screen.getByRole('button', { name: 'Применить' }))
    expect(await screen.findByText('export · export-0001')).toBeInTheDocument()
    expect(screen.queryByText('user · user-marketer-2')).not.toBeInTheDocument()
  })
})
