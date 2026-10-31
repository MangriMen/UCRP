import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { MemoryRouter, useLocation } from 'react-router'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { userKeys, type Me } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { UsersManager } from './users-manager'

const manager = mockUsers[1]
const readOnlyManager: Me = {
  ...manager,
  permissions: manager.permissions.filter((code) => code !== 'user.write'),
}

function Search() {
  return <span data-testid="search">{useLocation().search}</span>
}

function renderManager(sessionUser: Me = manager, url = '/admin/users') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  client.setQueryData(userKeys.me, sessionUser)
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[url]}>
        <UsersManager />
        <Search />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const rowOf = async (text: string) => (await screen.findByText(text)).closest('tr') as HTMLElement

describe('UsersManager', () => {
  it('does not let the current user deactivate themselves', async () => {
    renderManager()
    const ownRow = await rowOf('manager@ucrp.test')
    const otherRow = await rowOf('analyst@ucrp.test')
    expect(within(ownRow).getByRole('button', { name: 'Деактивировать' })).toBeDisabled()
    expect(within(otherRow).getByRole('button', { name: 'Деактивировать' })).toBeEnabled()
  })

  it('hides write actions without the user.write permission', async () => {
    renderManager(readOnlyManager)
    await screen.findByText('analyst@ucrp.test')
    expect(screen.queryByRole('button', { name: 'Изменить' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Восстановить доступ' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Деактивировать' })).not.toBeInTheDocument()
  })

  it('offers recovery only for active users with an invitable role', async () => {
    renderManager()
    const tenantRow = await rowOf('tenant@ucrp.test')
    const analystRow = await rowOf('analyst@ucrp.test')
    const inactiveRow = await rowOf('olga@ucrp.test')
    expect(within(tenantRow).getByRole('button', { name: 'Восстановить доступ' })).toBeEnabled()
    expect(within(analystRow).queryByRole('button', { name: 'Восстановить доступ' })).toBeNull()
    expect(within(inactiveRow).queryByRole('button', { name: 'Восстановить доступ' })).toBeNull()
  })

  it('sends a recovery link after confirmation', async () => {
    renderManager()
    const row = await rowOf('tenant@ucrp.test')
    await userEvent.click(within(row).getByRole('button', { name: 'Восстановить доступ' }))
    const dialog = await screen.findByRole('alertdialog')
    expect(within(dialog).getByText(/tenant@ucrp\.test/)).toBeInTheDocument()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Отправить ссылку' }))
    expect(await screen.findByText('Ссылка отправлена')).toBeInTheDocument()
  })

  it('shows the API error when recovery delivery is unavailable', async () => {
    server.use(
      http.post(`${API}/users/:id/recovery`, () =>
        apiError(503, 'invitation_unavailable', 'Доставка приглашений не настроена'),
      ),
    )
    renderManager()
    const row = await rowOf('tenant@ucrp.test')
    await userEvent.click(within(row).getByRole('button', { name: 'Восстановить доступ' }))
    const dialog = await screen.findByRole('alertdialog')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Отправить ссылку' }))
    expect(await screen.findByText('Доставка приглашений не настроена')).toBeInTheDocument()
    expect(screen.queryByText('Ссылка отправлена')).not.toBeInTheDocument()
  })

  it('renames a user', async () => {
    renderManager()
    const row = await rowOf('analyst@ucrp.test')
    await userEvent.click(within(row).getByRole('button', { name: 'Изменить' }))
    const dialog = await screen.findByRole('dialog')
    expect(within(dialog).getByLabelText('Email')).toBeDisabled()
    const name = within(dialog).getByLabelText('Имя')
    await userEvent.clear(name)
    await userEvent.type(name, 'Анна Петрова')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Сохранить' }))
    expect(await screen.findByText('Анна Петрова')).toBeInTheDocument()
  })

  it('keeps the search and role filters in the URL', async () => {
    renderManager(manager, '/admin/users?role=analyst')
    expect(await screen.findByText('analyst@ucrp.test')).toBeInTheDocument()
    expect(screen.queryByText('tenant@ucrp.test')).not.toBeInTheDocument()
    await userEvent.type(screen.getByLabelText('Поиск пользователей'), 'анна')
    await waitFor(() => expect(screen.getByTestId('search')).toHaveTextContent('q=%D0%B0'))
  })
})
