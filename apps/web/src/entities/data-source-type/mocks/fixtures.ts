import type { DataSourceType } from '../model/types'

export const mockDataSourceTypes: DataSourceType[] = [
  {
    id: 'ds-type-0001',
    code: 'csv_excel',
    name: 'Файл CSV / Excel',
    connectorKey: 'csv_excel',
    supportsFile: true,
    supportsPull: false,
    isActive: true,
  },
  {
    id: 'ds-type-0002',
    code: 'crm_api',
    name: 'CRM по API',
    connectorKey: 'crm_api',
    supportsFile: false,
    supportsPull: true,
    isActive: true,
  },
  {
    id: 'ds-type-0003',
    code: 'pos_api',
    name: 'Касса по API',
    connectorKey: 'pos_api',
    supportsFile: true,
    supportsPull: true,
    isActive: false,
  },
  {
    id: 'ds-type-0004',
    code: 'api',
    name: 'Произвольный API',
    connectorKey: 'generic_api',
    supportsFile: false,
    supportsPull: true,
    isActive: true,
  },
]
