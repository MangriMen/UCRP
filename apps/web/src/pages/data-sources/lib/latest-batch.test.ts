import { describe, expect, it } from 'vitest'
import type { ImportBatch } from '@/entities/import-batch'
import { latestBatchBySource } from './latest-batch'
import { sourceKindLabel } from './source-kind'

const batch = (id: string, dataSourceId: string, createdAt: string) =>
  ({ id, dataSourceId, createdAt }) as ImportBatch

describe('latestBatchBySource', () => {
  it('keeps the newest batch per source regardless of order', () => {
    const latest = latestBatchBySource([
      batch('a', 'ds-1', '2026-10-01T00:00:00Z'),
      batch('b', 'ds-1', '2026-10-03T00:00:00Z'),
      batch('c', 'ds-2', '2026-10-02T00:00:00Z'),
      batch('d', 'ds-1', '2026-10-02T00:00:00Z'),
    ])
    expect(latest.get('ds-1')?.id).toBe('b')
    expect(latest.get('ds-2')?.id).toBe('c')
  })
})

describe('sourceKindLabel', () => {
  it('maps known codes and falls back to the type name', () => {
    expect(sourceKindLabel('crm_api', 'CRM по API')).toBe('CRM')
    expect(sourceKindLabel('pos_api', undefined)).toBe('POS')
    expect(sourceKindLabel('api', undefined)).toBe('API')
    expect(sourceKindLabel('csv_excel', 'Файл CSV / Excel')).toBe('Файл CSV / Excel')
  })
})
