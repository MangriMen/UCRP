import type { RequestHandler } from 'msw'
import { contactLimitHandlers } from '@/entities/contact-limit/mocks'
import { dataSourceHandlers } from '@/entities/data-source/mocks'
import { exportHandlers } from '@/entities/export/mocks'
import { guestHandlers } from '@/entities/guest/mocks'
import { importHandlers } from '@/entities/import-batch/mocks'
import { mergeCandidateHandlers } from '@/entities/merge-candidate/mocks'
import { recommendationHandlers } from '@/entities/recommendation/mocks'
import { campaignHandlers } from '@/entities/campaign/mocks'
import { segmentHandlers } from '@/entities/segment/mocks'
import { dataSourceTypeHandlers } from '@/entities/data-source-type/mocks'
import { eventTypeHandlers } from '@/entities/event-type/mocks'
import { projectHandlers } from '@/entities/project/mocks'
import { tenantHandlers } from '@/entities/tenant/mocks'
import { tenantCategoryHandlers } from '@/entities/tenant-category/mocks'
import { unmergeBlockHandlers } from '@/entities/unmerge-block/mocks'
import { userHandlers } from '@/entities/user/mocks'

export const handlers: RequestHandler[] = [
  ...userHandlers,
  ...dataSourceHandlers,
  ...importHandlers,
  ...guestHandlers,
  ...mergeCandidateHandlers,
  ...unmergeBlockHandlers,
  ...segmentHandlers,
  ...recommendationHandlers,
  ...exportHandlers,
  ...contactLimitHandlers,
  ...campaignHandlers,
  ...tenantCategoryHandlers,
  ...dataSourceTypeHandlers,
  ...eventTypeHandlers,
  ...projectHandlers,
  ...tenantHandlers,
]
