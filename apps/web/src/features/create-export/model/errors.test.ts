import { describe, expect, it } from 'vitest'
import { ApiError } from '@/shared/api/client'
import { exportErrorMessage } from './errors'

describe('exportErrorMessage', () => {
  it('explains the contact limit conflict', () => {
    const error = new ApiError(409, { code: 'contact_limit_exceeded', message: 'x' })
    expect(exportErrorMessage(error)).toMatch(/лимитом контактов/)
  })

  it('leaves other errors to the generic banner', () => {
    expect(
      exportErrorMessage(new ApiError(500, { code: 'internal_error', message: 'x' })),
    ).toBeNull()
    expect(exportErrorMessage(new Error('x'))).toBeNull()
  })
})
