import type { RequestHandler } from 'msw'
import { userHandlers } from '@/entities/user/mocks'

export const handlers: RequestHandler[] = [...userHandlers]
