import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { userKeys } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { UsersManager } from './users-manager'

function renderManager(sessionUserId?: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const sessionUser = mockUsers.find((u) => u.userId === sessionUserId)
  if (sessionUser) client.setQueryData(userKeys.me, sessionUser)
  render(
    <QueryClientProvider client={client}>
      <UsersManager />
    </QueryClientProvider>,
  )
}

describe('UsersManager', () => {
  it('requires a tenant when the tenant role is chosen', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить пользователя' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Email'), 'new@ucrp.test')
    await userEvent.type(within(dialog).getByLabelText('Имя'), 'Новый арендатор')
    await userEvent.click(within(dialog).getByLabelText('Роль'))
    await userEvent.click(await screen.findByRole('option', { name: 'Арендатор' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(
      await within(dialog).findByText('Выберите арендатора для роли «Арендатор»'),
    ).toBeInTheDocument()
  })

  it('creates a user without a tenant for other roles', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить пользователя' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Email'), 'new@ucrp.test')
    await userEvent.type(within(dialog).getByLabelText('Имя'), 'Пётр Новый')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(await screen.findByText('Пётр Новый')).toBeInTheDocument()
  })

  it('flags an existing email', async () => {
    renderManager()
    await userEvent.click(await screen.findByRole('button', { name: 'Добавить пользователя' }))
    const dialog = screen.getByRole('dialog')
    await userEvent.type(within(dialog).getByLabelText('Email'), 'analyst@ucrp.test')
    await userEvent.type(within(dialog).getByLabelText('Имя'), 'Дубликат')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Создать' }))
    expect(
      await within(dialog).findByText('Пользователь с таким email уже существует'),
    ).toBeInTheDocument()
  })

  it('does not let the current user deactivate themselves', async () => {
    renderManager('user-manager')
    const ownRow = (await screen.findByText('manager@ucrp.test')).closest('tr') as HTMLElement
    const otherRow = (await screen.findByText('analyst@ucrp.test')).closest('tr') as HTMLElement
    expect(within(ownRow).getByRole('button', { name: 'Деактивировать' })).toBeDisabled()
    expect(within(otherRow).getByRole('button', { name: 'Деактивировать' })).toBeEnabled()
  })

  it('locks the active flag when editing the current user', async () => {
    renderManager('user-manager')
    const ownRow = (await screen.findByText('manager@ucrp.test')).closest('tr') as HTMLElement
    await userEvent.click(within(ownRow).getByRole('button', { name: 'Изменить' }))
    const dialog = await screen.findByRole('dialog')
    expect(within(dialog).getByRole('checkbox', { name: 'Активен' })).toHaveAttribute(
      'aria-disabled',
      'true',
    )
  })
})
