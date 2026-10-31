import type { RequestHandler } from 'msw'
import { accessHandlers } from '@/entities/access/mocks'
import { auditLogHandlers } from '@/entities/audit-log/mocks'
import { contactLimitHandlers } from '@/entities/contact-limit/mocks'
import { dashboardHandlers } from '@/entities/dashboard/mocks'
import { dataQualityHandlers } from '@/entities/data-quality/mocks'
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
import { tenantAggregatesHandlers } from '@/entities/tenant-aggregates/mocks'
import { mallHandlers } from '@/entities/mall/mocks'
import { tenantHandlers } from '@/entities/tenant/mocks'
import { tenantCategoryHandlers } from '@/entities/tenant-category/mocks'
import { mergeBlockHandlers } from '@/entities/merge-block/mocks'
import { userHandlers } from '@/entities/user/mocks'

export const handlers: RequestHandler[] = [
  ...userHandlers,
  ...dataSourceHandlers,
  ...importHandlers,
  ...guestHandlers,
  ...mergeCandidateHandlers,
  ...mergeBlockHandlers,
  ...segmentHandlers,
  ...recommendationHandlers,
  ...exportHandlers,
  ...contactLimitHandlers,
  ...campaignHandlers,
  ...auditLogHandlers,
  ...tenantCategoryHandlers,
  ...dataSourceTypeHandlers,
  ...eventTypeHandlers,
  ...projectHandlers,
  ...tenantHandlers,
  ...mallHandlers,
  ...dashboardHandlers,
  ...tenantAggregatesHandlers,
  ...dataQualityHandlers,
  ...accessHandlers,
]
