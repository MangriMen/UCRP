import { ApiError } from '@/shared/api/client'

export const isFieldForbidden = (error: unknown) =>
  error instanceof ApiError && error.body.code === 'forbidden_field'
