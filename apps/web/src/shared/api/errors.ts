import { ApiError, type ApiErrorCode } from './client'

export const hasErrorCode = (error: unknown, code: ApiErrorCode) =>
  error instanceof ApiError && error.body.code === code
