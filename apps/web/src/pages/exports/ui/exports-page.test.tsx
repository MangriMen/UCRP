import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { RoleGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { ExportsPage } from './exports-page'

function renderAs(userId: string) {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={['/exports']}>
        <Routes>
          <Route
            path="/exports"
            element={
              <RoleGuard>
                <ExportsPage />
              </RoleGuard>
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

describe('ExportsPage', () => {
  it('offers new export and CSV download to the marketer', async () => {
    renderAs('user-marketer')
    expect((await screen.findAllByRole('button', { name: 'Скачать CSV' })).length).toBeGreaterThan(
      0,
    )
    expect(screen.getByRole('button', { name: 'Экспортировать' })).toBeInTheDocument()
    expect(screen.queryByText('Формат')).not.toBeInTheDocument()
  })

  it.each(['user-analyst', 'user-manager'])('keeps %s read-only', async (userId) => {
    renderAs(userId)
    expect((await screen.findAllByRole('button', { name: 'Скачать CSV' })).length).toBeGreaterThan(
      0,
    )
    expect(screen.queryByRole('button', { name: 'Экспортировать' })).not.toBeInTheDocument()
  })
})
