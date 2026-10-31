import { http, HttpResponse } from 'msw'
import { mockDataSourceTypes } from '@/entities/data-source-type/mocks'
import { API, apiError, mockId, paged } from '@/shared/api/mocks/helpers'
import type { DataSource, DataSourceCreateRequest, DataSourceUpdateRequest } from '../model/types'
import { DATA_SOURCE_WITH_IMPORTS_ID, mockDataSources } from './fixtures'

const sources: DataSource[] = structuredClone(mockDataSources)

const SORT_VALUES: Record<string, (source: DataSource) => string | null> = {
  name: (s) => s.name,
  tenantName: (s) => s.tenantName ?? null,
  typeName: (s) => s.typeCode ?? null,
  lastImportAt: (s) => s.lastImportAt ?? null,
  createdAt: (s) => s.createdAt ?? null,
}

const compareBySort = (sort: string) => {
  const direction = sort.startsWith('-') ? -1 : 1
  const value = SORT_VALUES[sort.replace('-', '')] ?? SORT_VALUES.createdAt
  return (a: DataSource, b: DataSource) => {
    const left = value(a)
    const right = value(b)
    if (left === right) return 0
    if (left === null) return 1
    if (right === null) return -1
    return left < right ? -direction : direction
  }
}

export const dataSourceHandlers = [
  http.get(`${API}/data-sources`, ({ request }) => {
    const params = new URL(request.url).searchParams
    const tenantId = params.get('tenant_id')
    const typeId = params.get('type_id')
    const isActive = params.get('is_active')
    const importStatus = params.get('import_status')
    const q = params.get('q')?.toLowerCase()
    const filtered = sources
      .filter(
        (s) =>
          (!tenantId || s.tenantId === tenantId) &&
          (!typeId || s.typeId === typeId) &&
          (isActive === null || String(s.isActive) === isActive) &&
          (!importStatus || (s.lastImport?.status ?? 'none') === importStatus) &&
          (!q || `${s.name} ${s.tenantName ?? ''}`.toLowerCase().includes(q)),
      )
      .sort(compareBySort(params.get('sort') ?? '-createdAt'))
    return HttpResponse.json(paged(request, filtered))
  }),

  http.get(`${API}/data-sources/:id`, ({ params }) => {
    const source = sources.find((s) => s.id === params.id)
    if (!source) return apiError(404, 'not_found', 'Источник не найден')
    return HttpResponse.json(source)
  }),

  http.post(`${API}/data-sources/:id/sync`, ({ params }) => {
    const source = sources.find((s) => s.id === params.id)
    if (!source) return apiError(404, 'not_found', 'Источник не найден')
    const status = source.lastImport?.status
    if (status === 'queued' || status === 'running') {
      return apiError(409, 'sync_in_progress', 'Синхронизация уже выполняется')
    }
    const batchId = mockId('batch-new')
    source.lastImport = {
      batchId,
      status: 'queued',
      createdAt: new Date().toISOString(),
      finishedAt: null,
      rejectedRows: 0,
    }
    return HttpResponse.json({ batchId, status: 'queued' }, { status: 202 })
  }),

  http.post(`${API}/data-sources`, async ({ request }) => {
    const body = (await request.json()) as DataSourceCreateRequest
    const type = mockDataSourceTypes.find((t) => t.id === body.typeId)
    if (!type) return apiError(400, 'validation_error', 'Неизвестный тип источника')
    const now = new Date().toISOString()
    const source: DataSource = {
      ...body,
      id: mockId('ds-new'),
      typeCode: type.code,
      tenantName: null,
      isActive: body.isActive ?? true,
      lastImportAt: null,
      createdAt: now,
      updatedAt: now,
    }
    sources.push(source)
    return HttpResponse.json(source, { status: 201 })
  }),

  http.patch(`${API}/data-sources/:id`, async ({ params, request }) => {
    const source = sources.find((s) => s.id === params.id)
    if (!source) return apiError(404, 'not_found', 'Источник не найден')
    const body = (await request.json()) as DataSourceUpdateRequest
    Object.assign(source, body, { updatedAt: new Date().toISOString() })
    return HttpResponse.json(source)
  }),

  http.delete(`${API}/data-sources/:id`, ({ params }) => {
    const index = sources.findIndex((s) => s.id === params.id)
    if (index === -1) return apiError(404, 'not_found', 'Источник не найден')
    if (params.id === DATA_SOURCE_WITH_IMPORTS_ID) {
      return apiError(409, 'conflict', 'У источника есть загрузки, удаление невозможно')
    }
    sources.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
]
