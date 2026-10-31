import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it, vi } from 'vitest'
import { TenantAggregatesPage } from './tenant-aggregates-page'

const auth = vi.hoisted(() => ({ tenantId: 'tenant-0001' as string | null }))

vi.mock('@/entities/user', () => ({
  useAuth: () => ({
    user: { tenantId: auth.tenantId },
    role: 'tenant',
    can: (permission: string) =>
      ['metric.new_guests_via_mall.read', 'metric.repeat_visits.read'].includes(permission),
  }),
}))

function renderPage() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter>
        <TenantAggregatesPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('TenantAggregatesPage', () => {
  it('requests the session tenant and explains a suppressed count', async () => {
    auth.tenantId = 'tenant-0001'
    renderPage()
    expect(await screen.findByText('Часть значений скрыта')).toBeInTheDocument()
    expect(screen.getByTestId('visits-count')).toHaveTextContent('меньше 5')
    expect(screen.getByTestId('new-guests-count')).toHaveTextContent('меньше 5')
    expect(screen.getByTestId('repeat-visits-count')).toHaveTextContent('меньше 5')
    expect(screen.queryByTestId('revenue-total')).not.toBeInTheDocument()
  })

  it('does not query anything without a session tenant', () => {
    auth.tenantId = null
    renderPage()
    expect(screen.getByText('Арендатор не привязан к вашей учётной записи')).toBeInTheDocument()
  })
})
