import type { DataSource } from '../model/types'

export const mockDataSources: DataSource[] = [
  {
    id: 'ds-0001',
    tenantId: 'tenant-0001',
    tenantName: 'Лента',
    typeId: 'type-crm',
    typeCode: 'crm',
    name: 'CRM «Лента»',
    isActive: true,
    lastImportAt: '2026-10-28T09:01:10Z',
  },
  {
    id: 'ds-0002',
    tenantId: 'tenant-0002',
    tenantName: 'Зара',
    typeId: 'type-pos',
    typeCode: 'pos',
    name: 'Касса «Зара»',
    isActive: true,
    lastImportAt: '2026-10-29T11:31:00Z',
  },
  {
    id: 'ds-0003',
    tenantId: 'tenant-0003',
    tenantName: 'Спортмастер',
    typeId: 'type-api',
    typeCode: 'api',
    name: 'API «Спортмастер»',
    isActive: true,
    lastImportAt: null,
  },
]
