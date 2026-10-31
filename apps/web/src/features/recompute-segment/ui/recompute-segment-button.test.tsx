import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { RecomputeSegmentButton } from './recompute-segment-button'

function renderButton(status: 'ready' | 'computing') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <RecomputeSegmentButton segmentId="segment-0001" status={status} />
    </QueryClientProvider>,
  )
}

describe('RecomputeSegmentButton', () => {
  it('is disabled while the segment is computing', () => {
    renderButton('computing')
    expect(screen.getByRole('button', { name: 'Пересчёт…' })).toBeDisabled()
  })

  it('shows the conflict returned by the server', async () => {
    server.use(
      http.post(`${API}/segments/:id/recompute`, () =>
        apiError(409, 'recompute_in_progress', 'Пересчёт уже выполняется'),
      ),
    )
    renderButton('ready')
    await userEvent.click(screen.getByRole('button', { name: 'Пересчитать' }))
    expect(await screen.findByText('Пересчёт уже выполняется')).toBeInTheDocument()
  })
})
