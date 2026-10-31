import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { UsersManager } from './users-manager'

function renderManager() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
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
})
