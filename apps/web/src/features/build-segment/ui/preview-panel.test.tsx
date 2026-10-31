import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { http } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { PreviewPanel } from './preview-panel'

function renderPanel(rule: Parameters<typeof PreviewPanel>[0]['rule']) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <PreviewPanel rule={rule} />
    </QueryClientProvider>,
  )
}

describe('PreviewPanel', () => {
  it('shows the audience size after the debounce', async () => {
    renderPanel({ periodDays: 90 })
    expect(await screen.findByText('1200')).toBeInTheDocument()
  })

  it('does not request a preview for an invalid rule', () => {
    renderPanel(null)
    expect(screen.getByText('Заполните условия корректно')).toBeInTheDocument()
  })

  it('shows the timeout returned by the server', async () => {
    server.use(
      http.post(`${API}/segments/preview`, () => apiError(504, 'timeout', 'Слишком долго')),
    )
    renderPanel({ periodDays: 730 })
    expect(await screen.findByText('Сервер не ответил вовремя')).toBeInTheDocument()
  })
})
