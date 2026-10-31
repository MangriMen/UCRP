import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { describe, expect, it } from 'vitest'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { server } from '@/app/mocks/server'
import { RecommendationList } from './recommendation-list'

function renderList(guestId: string) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <RecommendationList guestId={guestId} />
    </QueryClientProvider>,
  )
}

describe('RecommendationList', () => {
  it('explains factors with value, weight, score and algorithm', async () => {
    renderList('guest-0001')
    await userEvent.click((await screen.findAllByRole('button', { name: 'Почему?' }))[0])
    expect(await screen.findByText('Похожие гости')).toBeInTheDocument()
    expect(screen.getByText('32%')).toBeInTheDocument()
    expect(screen.getByText('Значение 80% · вес 40%')).toBeInTheDocument()
    expect(screen.getByText('Оценка 0.78 · алгоритм v1')).toBeInTheDocument()
  })

  it('marks a fallback recommendation as based on little data', async () => {
    renderList('guest-0003')
    expect(await screen.findByText('Мало данных')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Почему?' }))
    expect(await screen.findByText(/Истории гостя мало/)).toBeInTheDocument()
    expect(screen.getByText('Популярность')).toBeInTheDocument()
  })

  it('shows the API error', async () => {
    server.use(
      http.get(`${API}/guests/:id/recommendations`, () =>
        apiError(403, 'forbidden_permission', 'Нет доступа'),
      ),
    )
    renderList('guest-0001')
    expect(await screen.findByText(/Нет доступа/)).toBeInTheDocument()
  })
})
