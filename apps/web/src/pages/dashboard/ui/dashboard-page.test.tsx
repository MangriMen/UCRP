import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { DashboardPage } from './dashboard-page'

function renderPage() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <DashboardPage />
    </QueryClientProvider>,
  )
}

describe('DashboardPage', () => {
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
})
