import type { DataSource } from '../model/types'

export const DATA_SOURCE_WITH_IMPORTS_ID = 'ds-0001'

export const mockDataSources: DataSource[] = [
  {
    id: 'ds-0001',
    tenantId: 'tenant-0001',
    tenantName: 'Лента',
    typeId: 'ds-type-0002',
    typeCode: 'crm_api',
    name: 'CRM «Лента»',
    fieldMapping: { externalId: 'id', occurredAt: 'date' },
    isActive: true,
    lastImportAt: '2026-10-28T09:01:10Z',
    lastImport: {
      batchId: 'batch-0004',
      status: 'running',
      createdAt: '2026-10-31T08:00:00Z',
      finishedAt: null,
      rejectedRows: 4,
    },
  },
  {
    id: 'ds-0002',
    tenantId: 'tenant-0002',
    tenantName: 'Зара',
    typeId: 'ds-type-0003',
    typeCode: 'pos_api',
    name: 'Касса «Зара»',
    fieldMapping: { externalId: 'check_id', occurredAt: 'date_time' },
    isActive: true,
    lastImportAt: '2026-10-29T11:31:00Z',
    lastImport: {
      batchId: 'batch-0002',
      status: 'completed_with_errors',
      createdAt: '2026-10-29T11:30:00Z',
      finishedAt: '2026-10-29T11:31:00Z',
      rejectedRows: 42,
    },
  },
  {
    id: 'ds-0003',
    tenantId: 'tenant-0003',
    tenantName: 'Спортмастер',
    typeId: 'ds-type-0004',
    typeCode: 'api',
    name: 'API «Спортмастер»',
    fieldMapping: { externalId: 'uid', occurredAt: 'ts' },
    isActive: true,
    lastImportAt: null,
    lastImport: {
      batchId: 'batch-0003',
      status: 'failed',
      createdAt: '2026-10-30T02:00:00Z',
      finishedAt: '2026-10-30T02:00:33Z',
      rejectedRows: 0,
    },
  },
]
