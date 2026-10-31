import type { DataQualitySource, DataQualitySourceDetail } from '../model/types'

export const mockQualitySources: DataQualitySource[] = [
  {
    dataSourceId: 'ds-0001',
    dataSourceName: 'CRM «Лента»',
    tenantId: 'tenant-0001',
    tenantName: 'Лента',
    lastSuccessfulImportAt: '2026-10-28T09:01:10Z',
    rejectedShare30d: 0.02,
    pendingMergeCandidates: 3,
    hasWarning: false,
    warningReasons: [],
  },
  {
    dataSourceId: 'ds-0002',
    dataSourceName: 'Касса «Зара»',
    tenantId: 'tenant-0002',
    tenantName: 'Зара',
    lastSuccessfulImportAt: '2026-10-29T11:31:00Z',
    rejectedShare30d: 0.12,
    pendingMergeCandidates: 14,
    hasWarning: true,
    warningReasons: ['rejected_share_above_10pct'],
  },
  {
    dataSourceId: 'ds-0003',
    dataSourceName: 'API «Спортмастер»',
    tenantId: 'tenant-0003',
    tenantName: 'Спортмастер',
    lastSuccessfulImportAt: null,
    rejectedShare30d: 0,
    pendingMergeCandidates: 0,
    hasWarning: true,
    warningReasons: ['data_older_than_7_days'],
  },
]

export const mockQualityDetails: Record<
  string,
  Pick<DataQualitySourceDetail, 'series' | 'recentBatches'>
> = {
  'ds-0001': {
    series: [
      { date: '2026-10-14', totalRows: 1000, rejectedRows: 30, rejectedShare: 0.03 },
      { date: '2026-10-28', totalRows: 1200, rejectedRows: 20, rejectedShare: 0.017 },
    ],
    recentBatches: [
      {
        id: 'batch-0001',
        dataSourceId: 'ds-0001',
        dataSourceName: 'CRM «Лента»',
        tenantId: 'tenant-0001',
        mode: 'file',
        fileName: 'crm-october.csv',
        status: 'completed',
        totalRows: 1200,
        acceptedRows: 1180,
        rejectedRows: 20,
        duplicateRows: 0,
        createdAt: '2026-10-28T09:00:00Z',
      },
    ],
  },
  'ds-0002': {
    series: [
      { date: '2026-10-22', totalRows: 500, rejectedRows: 70, rejectedShare: 0.14 },
      { date: '2026-10-29', totalRows: 540, rejectedRows: 42, rejectedShare: 0.078 },
    ],
    recentBatches: [
      {
        id: 'batch-0002',
        dataSourceId: 'ds-0002',
        dataSourceName: 'Касса «Зара»',
        tenantId: 'tenant-0002',
        mode: 'file',
        fileName: 'pos-week-43.xlsx',
        status: 'completed_with_errors',
        totalRows: 540,
        acceptedRows: 498,
        rejectedRows: 42,
        duplicateRows: 0,
        createdAt: '2026-10-29T11:30:00Z',
      },
    ],
  },
  'ds-0003': { series: [], recentBatches: [] },
}
