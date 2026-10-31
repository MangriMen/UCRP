import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { AdminDictionariesPage } from './admin-dictionaries-page'

const SESSION_KEY = 'ucrp-mock-session'
const [marketer, , , analyst] = mockUsers

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderPage(userId: string, url: string) {
  localStorage.setItem(SESSION_KEY, userId)
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route
            path="/admin/dictionaries"
            element={
              <>
                <AdminDictionariesPage />
                <Search />
              </>
            }
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('AdminDictionariesPage', () => {
  afterEach(() => {
    localStorage.clear()
    setAccessToken(null)
  })

  it('opens category relations from a deep link and clears the link on close', async () => {
    renderPage(analyst.userId, '/admin/dictionaries?relations=category-0001')
    const dialog = within(await screen.findByRole('dialog'))
    expect(dialog.getByText('Связи категории «Питание»')).toBeInTheDocument()
    expect(await dialog.findByRole('button', { name: 'Сохранить' })).toBeInTheDocument()
    await userEvent.click(dialog.getByRole('button', { name: 'Отмена' }))
    await waitFor(() => expect(screen.getByTestId('search')).toHaveTextContent(/^$/))
  })

  it('opens relations read-only without reference.write', async () => {
    renderPage(marketer.userId, '/admin/dictionaries')
    await screen.findByText('Питание')
    const row = within(screen.getByText('Питание').closest('tr') as HTMLElement)
    expect(row.queryByRole('button', { name: 'Изменить' })).not.toBeInTheDocument()
    await userEvent.click(row.getByRole('button', { name: 'Связи' }))
    expect(screen.getByTestId('search')).toHaveTextContent('?relations=category-0001')
    const dialog = within(await screen.findByRole('dialog'))
    expect(await dialog.findByText('Развлечения')).toBeInTheDocument()
    expect(dialog.queryByRole('button', { name: 'Сохранить' })).not.toBeInTheDocument()
  })
})
