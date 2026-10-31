import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { TenantAggregatesPage } from './tenant-aggregates-page'

const auth = vi.hoisted(() => ({ tenantId: 'tenant-0001' as string | null }))

vi.mock('@/entities/user', () => ({
  useAuth: () => ({ user: { tenantId: auth.tenantId }, role: 'tenant' }),
}))

function renderPage() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <TenantAggregatesPage />
    </QueryClientProvider>,
  )
}

describe('TenantAggregatesPage', () => {
  it('requests the session tenant and explains suppressed cells', async () => {
    auth.tenantId = 'tenant-0001'
    renderPage()
    expect(await screen.findByText('Часть значений скрыта')).toBeInTheDocument()
    expect(screen.getByTestId('new-guests')).toHaveTextContent('148')
    expect(screen.getByTestId('repeat-visits')).toHaveTextContent('меньше 5')
    expect(screen.getByText('Мой арендатор: Лента')).toBeInTheDocument()
  })

  it('does not query anything without a session tenant', () => {
    auth.tenantId = null
    renderPage()
    expect(screen.getByText('Арендатор не привязан к вашей учётной записи')).toBeInTheDocument()
  })
})
