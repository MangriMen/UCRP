import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { RoleGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { SegmentDetailsPage } from './segment-details-page'

function renderAs(userId: string) {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={['/segments/segment-0001']}>
        <Routes>
          <Route
            path="/segments/:id"
            element={
              <RoleGuard>
                <SegmentDetailsPage />
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

describe('SegmentDetailsPage', () => {
  it('lets the marketer edit, recompute and export', async () => {
    renderAs('user-marketer')
    expect(await screen.findByRole('button', { name: 'Изменить правило' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Пересчитать' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Экспортировать' })).toBeInTheDocument()
  })

  it('lets the analyst edit and recompute but not export', async () => {
    renderAs('user-analyst')
    expect(await screen.findByRole('button', { name: 'Изменить правило' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Пересчитать' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Экспортировать' })).not.toBeInTheDocument()
  })

  it('keeps the manager read-only', async () => {
    renderAs('user-manager')
    expect(await screen.findByRole('link', { name: 'К списку' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Изменить правило' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Пересчитать' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Экспортировать' })).not.toBeInTheDocument()
  })
})
