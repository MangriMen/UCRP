import type { ImportBatch } from './types'

export function latestBatchBySource(batches: ImportBatch[]): Map<string, ImportBatch> {
  const latest = new Map<string, ImportBatch>()
  for (const batch of batches) {
    const current = latest.get(batch.dataSourceId)
    if (!current || batch.createdAt > current.createdAt) latest.set(batch.dataSourceId, batch)
  }
  return latest
}
