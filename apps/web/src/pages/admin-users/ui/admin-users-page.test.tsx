import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http } from 'msw'
import { MemoryRouter } from 'react-router'
import { describe, expect, it } from 'vitest'
import { server } from '@/app/mocks/server'
import { userKeys, type Me } from '@/entities/user'
import { mockUsers } from '@/entities/user/mocks'
import { API, apiError } from '@/shared/api/mocks/helpers'
import { AdminUsersPage } from './admin-users-page'

const manager = mockUsers[1]

function renderPage(sessionUser: Me = manager) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  client.setQueryData(userKeys.me, sessionUser)
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={['/admin/users']}>
        <AdminUsersPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

async function openInviteDialog() {
  await userEvent.click(await screen.findByRole('button', { name: 'Пригласить пользователя' }))
  return screen.getByRole('dialog')
}

async function fillInvite(dialog: HTMLElement, email: string, name: string) {
  await userEvent.type(within(dialog).getByLabelText('Email'), email)
  await userEvent.type(within(dialog).getByLabelText('Имя'), name)
}

const submitInvite = (dialog: HTMLElement) =>
  userEvent.click(within(dialog).getByRole('button', { name: 'Отправить приглашение' }))

describe('AdminUsersPage', () => {
  it('hides the invitation without the user.write permission', async () => {
    renderPage({
      ...manager,
      permissions: manager.permissions.filter((code) => code !== 'user.write'),
    })
    expect(await screen.findByText('analyst@ucrp.test')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Пригласить пользователя' })).toBeNull()
  })

  it('offers only the roles the API can invite', async () => {
    renderPage()
    const dialog = await openInviteDialog()
    await userEvent.click(within(dialog).getByLabelText('Роль'))
    const options = await screen.findAllByRole('option')
    expect(options.map((option) => option.textContent)).toEqual(['Управляющий', 'Арендатор'])
  })

  it('requires a tenant for the tenant role', async () => {
    renderPage()
    const dialog = await openInviteDialog()
    await fillInvite(dialog, 'new-tenant@ucrp.test', 'Новый арендатор')
    await userEvent.click(within(dialog).getByLabelText('Роль'))
    await userEvent.click(await screen.findByRole('option', { name: 'Арендатор' }))
    await submitInvite(dialog)
    expect(
      await within(dialog).findByText('Выберите арендатора для роли «Арендатор»'),
    ).toBeInTheDocument()
  })

  it('confirms where the invitation was sent', async () => {
    renderPage()
    const dialog = await openInviteDialog()
    await fillInvite(dialog, 'new-manager@ucrp.test', 'Пётр Новый')
    await submitInvite(dialog)
    expect(await screen.findByText('Приглашение отправлено')).toBeInTheDocument()
    expect(screen.getByText(/new-manager@ucrp\.test/)).toBeInTheDocument()
  })

  it('flags an email that already has an account', async () => {
    renderPage()
    const dialog = await openInviteDialog()
    await fillInvite(dialog, 'analyst@ucrp.test', 'Дубликат')
    await submitInvite(dialog)
    expect(
      await within(dialog).findByText('Пользователь с таким email уже есть'),
    ).toBeInTheDocument()
  })

  it('explains that invitation delivery is not configured', async () => {
    server.use(
      http.post(`${API}/users/invitations`, () =>
        apiError(503, 'invitation_unavailable', 'Доставка приглашений не настроена'),
      ),
    )
    renderPage()
    const dialog = await openInviteDialog()
    await fillInvite(dialog, 'new-manager@ucrp.test', 'Пётр Новый')
    await submitInvite(dialog)
    expect(await within(dialog).findByText('Доставка приглашений не настроена')).toBeInTheDocument()
  })
})
