import type { RequestHandler } from 'msw'
import { dataSourceHandlers } from '@/entities/data-source/mocks'
import { guestHandlers } from '@/entities/guest/mocks'
import { importHandlers } from '@/entities/import-batch/mocks'
import { userHandlers } from '@/entities/user/mocks'

export const handlers: RequestHandler[] = [
  ...userHandlers,
  ...dataSourceHandlers,
  ...importHandlers,
  ...guestHandlers,
]
