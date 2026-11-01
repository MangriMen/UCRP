import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, useLocation } from 'react-router'
import { describe, expect, it } from 'vitest'
import { userKeys } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { AccessEffectivePage } from './access-effective-page'

const manager = mockUsers[1]

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderPage(url = '/access/effective') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  client.setQueryData(userKeys.me, manager)
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[url]}>
        <AccessEffectivePage />
        <Search />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('AccessEffectivePage', () => {
  it('asks for a subject first', async () => {
    renderPage()
    expect(await screen.findByText('Выберите субъекта')).toBeInTheDocument()
  })

  it('shows the effective permissions of the subject from the URL', async () => {
    renderPage('/access/effective?subject_type=tenant&subject_id=tenant-0001')
    expect(await screen.findByText('Агрегаты только своего арендатора')).toBeInTheDocument()
    expect(screen.getByText('Справочники')).toBeInTheDocument()
    expect(screen.getByText(/Заря: прав 5/)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Гранты субъекта' })).toHaveAttribute(
      'href',
      '/access/grants?subject_type=tenant&subject_id=tenant-0001',
    )
  })

  it('writes the chosen subject to the URL', async () => {
    renderPage()
    await userEvent.click(await screen.findByLabelText('Тип субъекта'))
    await userEvent.click(await screen.findByRole('option', { name: 'Все арендаторы (роль)' }))
    expect(screen.getByTestId('search')).toHaveTextContent('subject_type=role&subject_id=tenant')
    expect(await screen.findByText('Справочники')).toBeInTheDocument()
  })
})
