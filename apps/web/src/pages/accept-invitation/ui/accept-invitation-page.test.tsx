import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'
import { MOCK_INVITATION_TOKEN } from '@/entities/user/mocks'
import { AcceptInvitationPage } from './accept-invitation-page'

function renderPage(url: string) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter initialEntries={[url]}>
        <Routes>
          <Route path="/invite/accept" element={<AcceptInvitationPage />} />
          <Route path="/login" element={<h1>Вход</h1>} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

async function setPassword(password: string, confirmation = password) {
  await userEvent.type(screen.getByLabelText('Новый пароль'), password)
  await userEvent.type(screen.getByLabelText('Повторите пароль'), confirmation)
  await userEvent.click(screen.getByRole('button', { name: 'Установить пароль' }))
}

describe('AcceptInvitationPage', () => {
  it('explains a link without a token', () => {
    renderPage('/invite/accept')
    expect(screen.getByText('Ссылка неполная')).toBeInTheDocument()
    expect(screen.queryByLabelText('Новый пароль')).toBeNull()
  })

  it('validates the password before sending it', async () => {
    renderPage(`/invite/accept?token=${MOCK_INVITATION_TOKEN}`)
    await setPassword('short')
    expect(await screen.findByText('Не менее 8 символов')).toBeInTheDocument()
  })

  it('requires matching passwords', async () => {
    renderPage(`/invite/accept?token=${MOCK_INVITATION_TOKEN}`)
    await setPassword('long-password-1', 'long-password-2')
    expect(await screen.findByText('Пароли не совпадают')).toBeInTheDocument()
  })

  it('reports an expired, used or unknown link', async () => {
    renderPage('/invite/accept?token=expired-token')
    await setPassword('long-password-1')
    expect(await screen.findByText('Ссылка недействительна')).toBeInTheDocument()
  })

  it('sends the user to the login page after setting the password', async () => {
    renderPage(`/invite/accept?token=${MOCK_INVITATION_TOKEN}`)
    await setPassword('long-password-1')
    expect(await screen.findByText('Пароль установлен')).toBeInTheDocument()
    await userEvent.click(screen.getByText('Перейти ко входу'))
    expect(await screen.findByRole('heading', { name: 'Вход' })).toBeInTheDocument()
  })
})
