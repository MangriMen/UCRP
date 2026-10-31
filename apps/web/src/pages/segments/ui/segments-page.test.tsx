import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { AccessGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { SegmentsPage } from './segments-page'

function renderAs(userId: string) {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={['/segments']}>
        <Routes>
          <Route
            path="/segments"
            element={
              <AccessGuard>
                <SegmentsPage />
              </AccessGuard>
            }
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

afterEach(() => {
  localStorage.clear()
  setAccessToken(null)
})

describe('SegmentsPage', () => {
  it.each([
    ['user-marketer', true],
    ['user-analyst', true],
    ['user-manager', false],
  ])('%s can create a segment: %s', async (userId, allowed) => {
    renderAs(userId)
    expect(await screen.findByText('Спящий сосед')).toBeInTheDocument()
    expect(!!screen.queryByRole('button', { name: 'Новый сегмент' })).toBe(allowed)
  })
})
