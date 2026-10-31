import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { describe, expect, it, vi } from 'vitest'
import { DataSourcesPage } from './data-sources-page'

vi.mock('@/entities/user', () => ({ useAuth: () => ({ role: 'analyst' }) }))

function renderPage() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter>
        <DataSourcesPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const rowOf = (name: string) => screen.getByText(name).closest('tr') as HTMLElement

describe('DataSourcesPage', () => {
  it('shows the last import status and sync only for pull-capable types', async () => {
    renderPage()
    await screen.findByText('Касса «Зара»')
    expect(within(rowOf('API «Спортмастер»')).getByText('Ошибка')).toBeInTheDocument()
    expect(within(rowOf('Касса «Зара»')).getByText('Завершён с ошибками')).toBeInTheDocument()
    expect(
      within(rowOf('Касса «Зара»')).getByRole('button', { name: 'Синхронизировать' }),
    ).toBeInTheDocument()
    expect(
      within(rowOf('Касса «Зара»')).getByRole('link', { name: 'Загрузить файл' }),
    ).toHaveAttribute('href', '/import?source=ds-0002')
    expect(within(rowOf('CRM «Лента»')).queryByRole('link')).not.toBeInTheDocument()
  })

  it('filters by tenant search', async () => {
    renderPage()
    await screen.findByText('Касса «Зара»')
    await userEvent.type(screen.getByLabelText('Поиск по арендатору'), 'спорт')
    expect(screen.queryByText('Касса «Зара»')).not.toBeInTheDocument()
    expect(screen.getByText('API «Спортмастер»')).toBeInTheDocument()
  })

  it('starts a sync and locks the button', async () => {
    renderPage()
    await screen.findByText('Касса «Зара»')
    const row = rowOf('API «Спортмастер»')
    await userEvent.click(within(row).getByRole('button', { name: 'Синхронизировать' }))
    expect(await within(row).findByRole('button', { name: 'Синхронизация…' })).toBeDisabled()
  })
})
