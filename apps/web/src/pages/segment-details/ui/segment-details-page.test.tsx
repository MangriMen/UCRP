import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { AccessGuard } from '@/entities/user'
import { setAccessToken } from '@/shared/api/client'
import { SegmentDetailsPage } from './segment-details-page'

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderAs(userId: string, url = '/segments/segment-0001') {
  localStorage.setItem('ucrp-mock-session', userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/segments/:id"
            element={
              <AccessGuard>
                <SegmentDetailsPage />
                <Search />
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

describe('SegmentDetailsPage', () => {
  it('lets the marketer edit, recompute and export', async () => {
    renderAs('user-marketer')
    expect(await screen.findByRole('button', { name: 'Изменить правило' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Пересчитать' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Выгрузить' })).toBeInTheDocument()
  })

  it('lets the analyst edit and recompute but not export', async () => {
    renderAs('user-analyst')
    expect(await screen.findByRole('button', { name: 'Изменить правило' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Пересчитать' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Выгрузить' })).not.toBeInTheDocument()
  })

  it('keeps the manager read-only', async () => {
    renderAs('user-manager')
    expect(await screen.findByRole('link', { name: 'Сегменты' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Изменить правило' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Пересчитать' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Выгрузить' })).not.toBeInTheDocument()
  })

  it('shows top tenants with category names and hides small groups', async () => {
    renderAs('user-marketer')
    const card = (await screen.findByText('Кого приглашать')).closest('[data-slot="card"]')
    expect(card).not.toBeNull()
    const scope = within(card as HTMLElement)
    expect(await scope.findByText('Кинотеатр «Синема»')).toBeInTheDocument()
    expect(scope.getByText('Развлечения')).toBeInTheDocument()
    expect(scope.getByText('64 из 110')).toBeInTheDocument()
    expect(scope.getByText(/Часть значений скрыта/)).toBeInTheDocument()
    expect(scope.getAllByText('скрыто').length).toBeGreaterThan(0)
  })

  it('keeps the top tenants limit in the URL', async () => {
    renderAs('user-marketer', '/segments/segment-0001?top=20')
    const select = await screen.findByRole('combobox', { name: 'Сколько арендаторов показать' })
    expect(select).toHaveTextContent('Топ 20')
    await userEvent.click(select)
    await userEvent.click(await screen.findByRole('option', { name: 'Топ 50' }))
    expect(screen.getByTestId('search')).toHaveTextContent('top=50')
  })

  it('hides top tenants from roles without recommendations', async () => {
    renderAs('user-manager')
    expect(await screen.findByRole('link', { name: 'Сегменты' })).toBeInTheDocument()
    expect(screen.queryByText('Кого приглашать')).not.toBeInTheDocument()
  })
})
