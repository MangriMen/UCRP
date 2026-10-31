import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { DataQualityPage } from './data-quality-page'

function renderPage() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter>
        <DataQualityPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const rowOf = (name: string) => screen.getByText(name).closest('tr') as HTMLElement

describe('DataQualityPage', () => {
  it('flags sources with warnings and explains the reasons', async () => {
    renderPage()
    await screen.findByText('Касса «Зара»')
    expect(within(rowOf('Касса «Зара»')).getByText('Требует внимания')).toBeInTheDocument()
    expect(within(rowOf('CRM «Лента»')).getByText('В норме')).toBeInTheDocument()
  })

  it('opens source details with recent imports and closes them', async () => {
    renderPage()
    await screen.findByText('Касса «Зара»')
    await userEvent.click(within(rowOf('Касса «Зара»')).getByRole('button', { name: 'Подробнее' }))
    expect(await screen.findByRole('link', { name: 'Открыть загрузку' })).toHaveAttribute(
      'href',
      '/import/batch-0002',
    )
    await userEvent.click(screen.getByRole('button', { name: 'Закрыть' }))
    expect(screen.queryByRole('link', { name: 'Открыть загрузку' })).not.toBeInTheDocument()
  })

  it('treats a source without imports as an empty state', async () => {
    renderPage()
    await screen.findByText('API «Спортмастер»')
    await userEvent.click(
      within(rowOf('API «Спортмастер»')).getByRole('button', { name: 'Подробнее' }),
    )
    expect(await screen.findByText('Загрузок пока не было')).toBeInTheDocument()
  })
})
