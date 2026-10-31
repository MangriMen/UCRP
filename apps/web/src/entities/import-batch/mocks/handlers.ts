import { http, HttpResponse } from 'msw'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type { ImportBatch } from '../model/types'
import { mockBatches, mockDataSources, mockRejectedRows } from './fixtures'

const MAX_FILE_BYTES = 10 * 1024 * 1024
const ACCEPTED_EXTENSIONS = ['.csv', '.xlsx']
const TOTAL_ROWS = 300

const batches: ImportBatch[] = structuredClone(mockBatches)

function advance(batch: ImportBatch) {
  if (batch.status === 'queued') {
    batch.status = 'running'
    batch.totalRows = TOTAL_ROWS
    batch.startedAt = new Date().toISOString()
  } else if (batch.status === 'running') {
    const step = batch.totalRows / 3
    batch.acceptedRows = Math.min(batch.totalRows, batch.acceptedRows + step)
    if (batch.acceptedRows + batch.rejectedRows + batch.duplicateRows >= batch.totalRows) {
      batch.status = 'completed'
      batch.finishedAt = new Date().toISOString()
    }
  }
}

function enqueue(dataSourceId: string, fields: Partial<ImportBatch>): ImportBatch {
  const source = mockDataSources.find((s) => s.id === dataSourceId)
  const batch: ImportBatch = {
    id: mockId('batch-new'),
    dataSourceId,
    dataSourceName: source?.name,
    tenantId: source?.tenantId,
    mode: 'file',
    status: 'queued',
    totalRows: 0,
    acceptedRows: 0,
    rejectedRows: 0,
    duplicateRows: 0,
    createdAt: new Date().toISOString(),
    ...fields,
  }
  batches.unshift(batch)
  return batch
}

export const importHandlers = [
  http.get(`${API}/imports`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const dataSourceId = params.get('data_source_id')
    const status = params.get('status')
    const filtered = batches.filter(
      (b) => (!dataSourceId || b.dataSourceId === dataSourceId) && (!status || b.status === status),
    )
    filtered.forEach(advance)
    return HttpResponse.json(paged(request, filtered))
  }),

  http.post(`${API}/imports`, async ({ request }) => {
    const form = await request.formData()
    const dataSourceId = form.get('data_source_id')
    const file = form.get('file')
    if (typeof dataSourceId !== 'string' || file === null || typeof file === 'string') {
      return apiError(400, 'validation_error', 'Нужны источник и файл')
    }
    if (!mockDataSources.some((s) => s.id === dataSourceId)) {
      return apiError(404, 'not_found', 'Источник не найден')
    }
    if (!ACCEPTED_EXTENSIONS.some((ext) => file.name.toLowerCase().endsWith(ext))) {
      return apiError(422, 'bad_format', 'Поддерживаются только CSV и Excel')
    }
    if (file.size > MAX_FILE_BYTES) {
      return apiError(413, 'file_too_large', 'Файл больше 10 МБ')
    }
    const existing = batches.find(
      (b) => b.dataSourceId === dataSourceId && b.fileName === file.name,
    )
    if (existing) {
      return HttpResponse.json(
        { batchId: existing.id, status: existing.status, duplicate: true },
        { status: 200 },
      )
    }
    const batch = enqueue(dataSourceId, { fileName: file.name, fileHash: mockId('hash') })
    return HttpResponse.json(
      { batchId: batch.id, status: batch.status, duplicate: false, createdAt: batch.createdAt },
      { status: 202 },
    )
  }),

  http.get(`${API}/imports/:id`, ({ params }) => {
    const batch = batches.find((b) => b.id === params.id)
    if (!batch) return apiError(404, 'not_found', 'Импорт не найден')
    advance(batch)
    return HttpResponse.json(batch)
  }),

  http.get(`${API}/imports/:id/rejected-rows`, ({ params, request }) => {
    const batch = batches.find((b) => b.id === params.id)
    if (!batch) return apiError(404, 'not_found', 'Импорт не найден')
    return HttpResponse.json(paged(request, mockRejectedRows(batch)))
  }),
]
