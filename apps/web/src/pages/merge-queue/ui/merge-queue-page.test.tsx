import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { MergeQueuePage } from './merge-queue-page'

function renderPage(url = '/identity/queue') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/identity/queue" element={<MergeQueuePage />} />
          <Route path="/identity/queue/:id" element={<div>Сравнение профилей</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('MergeQueuePage', () => {
  it('shows pending candidates by default', async () => {
    renderPage()
    expect(await screen.findByText('Общий телефон, имена различаются')).toBeInTheDocument()
    expect(screen.queryByText('Общий email')).not.toBeInTheDocument()
  })

  it('shows every candidate when the status filter is cleared in the URL', async () => {
    renderPage('/identity/queue?resolution=all')
    expect(await screen.findByText('Общий email')).toBeInTheDocument()
  })

  it('opens the comparison on row click', async () => {
    renderPage()
    await userEvent.click(await screen.findByText('Общий телефон, имена различаются'))
    expect(await screen.findByText('Сравнение профилей')).toBeInTheDocument()
  })
})
