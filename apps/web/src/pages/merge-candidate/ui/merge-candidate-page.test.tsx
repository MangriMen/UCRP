import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { MergeCandidatePage } from './merge-candidate-page'

function renderPage(id: string) {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[`/identity/merge-candidates/${id}`]}>
        <Routes>
          <Route path="/identity/merge-candidates/:id" element={<MergeCandidatePage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('MergeCandidatePage', () => {
  it('compares both profiles side by side', async () => {
    renderPage('candidate-0001')
    expect(await screen.findByText(/Профиль A: Иван Петров/)).toBeInTheDocument()
    expect(screen.getByText(/Профиль B: Мария Иванова/)).toBeInTheDocument()
    expect(screen.getByText('Телефон: 79001234567', { selector: 'span' })).toBeInTheDocument()
  })

  it('shows an error for an unknown candidate', async () => {
    renderPage('missing')
    expect(await screen.findByText('Кандидат не найден')).toBeInTheDocument()
  })
})
