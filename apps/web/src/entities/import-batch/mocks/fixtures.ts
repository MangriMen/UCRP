import type { ImportBatch, RejectedRow } from '../model/types'

export const mockDataSources = [
  { id: 'ds-0001', name: 'CRM «Лента»', tenantId: 'tenant-0001' },
  { id: 'ds-0002', name: 'Касса «Зара»', tenantId: 'tenant-0002' },
  { id: 'ds-0003', name: 'API «Спортмастер»', tenantId: 'tenant-0003' },
]

const sourceFields = (dataSourceId: string) => {
  const source = mockDataSources.find((s) => s.id === dataSourceId) ?? mockDataSources[0]
  return { dataSourceId: source.id, dataSourceName: source.name, tenantId: source.tenantId }
}

export const mockBatches: ImportBatch[] = [
  {
    id: 'batch-0001',
    ...sourceFields('ds-0001'),
    mode: 'file',
    fileName: 'crm-october.csv',
    fileHash: 'a1b2c3d4',
    status: 'completed',
    totalRows: 1200,
    acceptedRows: 1180,
    rejectedRows: 0,
    duplicateRows: 20,
    createdAt: '2026-10-28T09:00:00Z',
    startedAt: '2026-10-28T09:00:05Z',
    finishedAt: '2026-10-28T09:01:10Z',
  },
  {
    id: 'batch-0002',
    ...sourceFields('ds-0002'),
    mode: 'file',
    fileName: 'pos-week-43.xlsx',
    fileHash: 'e5f6a7b8',
    status: 'completed_with_errors',
    totalRows: 540,
    acceptedRows: 498,
    rejectedRows: 42,
    duplicateRows: 0,
    createdAt: '2026-10-29T11:30:00Z',
    startedAt: '2026-10-29T11:30:04Z',
    finishedAt: '2026-10-29T11:31:00Z',
  },
  {
    id: 'batch-0003',
    ...sourceFields('ds-0003'),
    mode: 'pull',
    status: 'failed',
    totalRows: 0,
    acceptedRows: 0,
    rejectedRows: 0,
    duplicateRows: 0,
    errorMessage: 'Источник не ответил за 30 секунд',
    windowFrom: '2026-10-29T00:00:00Z',
    windowTo: '2026-10-30T00:00:00Z',
    createdAt: '2026-10-30T02:00:00Z',
    startedAt: '2026-10-30T02:00:03Z',
    finishedAt: '2026-10-30T02:00:33Z',
  },
  {
    id: 'batch-0004',
    ...sourceFields('ds-0001'),
    mode: 'file',
    fileName: 'crm-november.csv',
    fileHash: '99aa77bb',
    status: 'running',
    totalRows: 800,
    acceptedRows: 310,
    rejectedRows: 4,
    duplicateRows: 6,
    createdAt: '2026-10-31T08:00:00Z',
    startedAt: '2026-10-31T08:00:02Z',
  },
]

const rejectedTemplates: Pick<RejectedRow, 'errorCode' | 'errorMessage' | 'rawData'>[] = [
  { errorCode: 'bad_date', errorMessage: 'Дата не распознана', rawData: { date: '31.31.2026' } },
  { errorCode: 'no_external_id', errorMessage: 'Пустой идентификатор', rawData: { id: '' } },
  { errorCode: 'bad_amount', errorMessage: 'Сумма не число', rawData: { amount: 'abc' } },
  { errorCode: 'unknown_event_type', errorMessage: 'Тип «refund2» неизвестен', rawData: null },
  { errorCode: 'bad_phone', errorMessage: 'Телефон слишком короткий', rawData: { phone: '123' } },
]

export function mockRejectedRows(batch: ImportBatch): RejectedRow[] {
  return Array.from({ length: batch.rejectedRows }, (_, index) => ({
    id: `${batch.id}-row-${index + 1}`,
    importBatchId: batch.id,
    rowNumber: (index + 1) * 7,
    ...rejectedTemplates[index % rejectedTemplates.length],
  }))
}
