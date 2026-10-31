import type { RequestHandler } from 'msw'
import { importHandlers } from '@/entities/import-batch/mocks'
import { userHandlers } from '@/entities/user/mocks'

export const handlers: RequestHandler[] = [...userHandlers, ...importHandlers]
