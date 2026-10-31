import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ApiError } from '@/shared/api/client'
import { ErrorBanner } from './error-banner'

describe('ErrorBanner', () => {
  it('shows backend message, field details and request id', () => {
    const error = new ApiError(400, {
      code: 'validation_error',
      message: 'Неверные данные',
      details: [{ field: 'email', reason: 'invalid' }],
      request_id: 'req-1',
    })
    render(<ErrorBanner error={error} />)
    expect(screen.getByText('Проверьте введённые данные')).toBeInTheDocument()
    expect(screen.getByText('email: invalid')).toBeInTheDocument()
    expect(screen.getByText(/req-1/)).toBeInTheDocument()
  })

  it('shows a generic message for non-API errors', () => {
    render(<ErrorBanner error={new Error('boom')} />)
    expect(screen.getByText('Что-то пошло не так')).toBeInTheDocument()
  })
})
