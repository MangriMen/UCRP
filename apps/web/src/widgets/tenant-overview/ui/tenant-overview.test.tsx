import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { describe, expect, it, vi } from 'vitest'
import { TenantOverview } from './tenant-overview'

const auth = vi.hoisted(() => ({ permissions: [] as string[] }))

vi.mock('@/entities/user', () => ({
  useAuth: () => ({ can: (permission: string) => auth.permissions.includes(permission) }),
}))

const ALL_METRICS = [
  'metric.visits_count.read',
  'metric.new_guests_via_mall.read',
  'metric.repeat_visits.read',
  'metric.first_purchase_new_tenant.read',
  'metric.revenue_total.read',
  'metric.avg_check.read',
]

function renderOverview(tenantId: string, url = '/?from=2026-10-01&to=2026-10-07') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <TenantOverview tenantId={tenantId} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('TenantOverview', () => {
  it('shows every granted metric and marks a suppressed amount', async () => {
    auth.permissions = ALL_METRICS
    renderOverview('tenant-0002')
    expect(await screen.findByTestId('repeat-visits-count')).toHaveTextContent('82')
    expect(screen.getByTestId('first-purchases-count')).toHaveTextContent('17')
    expect(screen.getByTestId('revenue-total')).toHaveTextContent('45 120,00')
    expect(screen.getByTestId('avg-check')).toHaveTextContent('меньше 5 гостей')
    expect(screen.getByText('Часть значений скрыта')).toBeInTheDocument()
    expect(screen.getByRole('img', { name: /Выручка по дням: 7 дн\./ })).toBeInTheDocument()
  })

  it('hides metrics without their permission', async () => {
    auth.permissions = ['metric.visits_count.read']
    renderOverview('tenant-0002')
    expect(await screen.findByTestId('visits-count')).toHaveTextContent('130')
    for (const id of [
      'new-guests-count',
      'repeat-visits-count',
      'first-purchases-count',
      'revenue-total',
      'avg-check',
    ]) {
      expect(screen.queryByTestId(id)).not.toBeInTheDocument()
    }
    expect(screen.queryByRole('img', { name: /Выручка по дням/ })).not.toBeInTheDocument()
  })
})
