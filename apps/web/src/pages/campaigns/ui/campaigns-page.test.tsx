import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { CampaignsPage } from './campaigns-page'

function renderPage(url = '/campaigns') {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/campaigns" element={<CampaignsPage />} />
          <Route path="/campaigns/:id" element={<div>Карточка кампании</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('CampaignsPage', () => {
  it('restores the status filter from the URL', async () => {
    renderPage('/campaigns?status=draft')
    expect(await screen.findByText('Удержание клиентов')).toBeInTheDocument()
    expect(screen.queryByText('Высокий чек в «Заре»')).not.toBeInTheDocument()
  })

  it('opens the campaign on row click', async () => {
    renderPage()
    await userEvent.click(await screen.findByText('Высокий чек в «Заре»'))
    expect(await screen.findByText('Карточка кампании')).toBeInTheDocument()
  })
})
