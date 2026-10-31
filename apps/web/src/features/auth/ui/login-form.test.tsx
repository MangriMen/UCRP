import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { MOCK_PASSWORD, mockUsers } from '@/entities/user/mocks'
import { setAccessToken } from '@/shared/api/client'
import { LoginForm } from './login-form'

function renderForm(onSuccess = vi.fn()) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <LoginForm onSuccess={onSuccess} />
    </QueryClientProvider>,
  )
  return onSuccess
}

afterEach(() => {
  localStorage.clear()
  setAccessToken(null)
})

describe('LoginForm', () => {
  it('validates required fields', async () => {
    renderForm()
    await userEvent.click(screen.getByRole('button', { name: 'Войти' }))
    expect(await screen.findByText('Введите корректный email')).toBeInTheDocument()
  })

  it('shows the API error for wrong credentials', async () => {
    renderForm()
    await userEvent.type(screen.getByLabelText('Email'), mockUsers[0].email)
    await userEvent.type(screen.getByLabelText('Пароль'), 'wrong')
    await userEvent.click(screen.getByRole('button', { name: 'Войти' }))
    expect(await screen.findByText('Неверный email или пароль')).toBeInTheDocument()
  })

  it('calls onSuccess after a valid login', async () => {
    const onSuccess = renderForm()
    await userEvent.type(screen.getByLabelText('Email'), mockUsers[2].email)
    await userEvent.type(screen.getByLabelText('Пароль'), MOCK_PASSWORD)
    await userEvent.click(screen.getByRole('button', { name: 'Войти' }))
    await vi.waitFor(() => expect(onSuccess).toHaveBeenCalled())
  })
})
