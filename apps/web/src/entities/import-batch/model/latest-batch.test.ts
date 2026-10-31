import { describe, expect, it } from 'vitest'
import { latestBatchBySource } from './latest-batch'
import type { ImportBatch } from './types'

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
