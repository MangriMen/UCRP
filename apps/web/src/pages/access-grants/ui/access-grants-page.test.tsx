import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { describe, expect, it } from 'vitest'
import { userKeys, type Me } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { AccessGrantsPage } from './access-grants-page'

const manager = mockUsers[1]
const reader: Me = { ...manager, permissions: ['access.grant.read'] }

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderPage(url = '/access/grants', me: Me = manager) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  client.setQueryData(userKeys.me, me)
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/access/grants"
            element={
              <>
                <AccessGrantsPage />
                <Search />
              </>
            }
          />
          <Route path="/access/grants/:id" element={<div>Карточка гранта</div>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const rowOf = (text: string) => within(screen.getByText(text).closest('tr') as HTMLElement)

describe('AccessGrantsPage', () => {
  it('lists grants with subject names, content and status', async () => {
    renderPage()
    expect(await screen.findByText('Заря')).toBeInTheDocument()
    expect(rowOf('Заря').getByText('Действует')).toBeInTheDocument()
    expect(rowOf('Заря').getByText('Свои агрегаты')).toBeInTheDocument()
    expect(rowOf('Бистро «Уголок»').getByText('Черновик')).toBeInTheDocument()
    expect(rowOf('Все арендаторы').getByText('Роль')).toBeInTheDocument()
  })

  it('restores the status filter from the URL and resets it', async () => {
    renderPage('/access/grants?status=draft')
    expect(await screen.findByText('Бистро «Уголок»')).toBeInTheDocument()
    expect(screen.queryByText('Заря')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Сбросить' }))
    expect(await screen.findByText('Заря')).toBeInTheDocument()
  })

  it('filters by subject and keeps the choice in the URL', async () => {
    renderPage('/access/grants?subject_type=tenant')
    await userEvent.click(await screen.findByLabelText('Субъект'))
    await userEvent.click(await screen.findByRole('option', { name: 'Киномир' }))
    expect(screen.getByTestId('search')).toHaveTextContent('subject_id=tenant-0003')
    await waitFor(() => expect(screen.getAllByRole('row')).toHaveLength(2))
    expect(screen.getByText('Отозван')).toBeInTheDocument()
  })

  it('opens the grant card on row click', async () => {
    renderPage()
    await userEvent.click(await screen.findByText('Заря'))
    expect(await screen.findByText('Карточка гранта')).toBeInTheDocument()
  })

  it('offers grant creation only with access.grant.write', async () => {
    renderPage()
    expect(await screen.findByRole('button', { name: 'Выдать доступ' })).toBeInTheDocument()
  })

  it('hides grant creation from a read-only holder', async () => {
    renderPage('/access/grants', reader)
    await screen.findByText('Заря')
    expect(screen.queryByRole('button', { name: 'Выдать доступ' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Аудит доступа' })).not.toBeInTheDocument()
  })
})
