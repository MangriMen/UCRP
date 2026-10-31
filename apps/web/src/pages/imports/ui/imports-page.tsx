import { useNavigate, useSearchParams } from 'react-router'
import { useDataSources } from '@/entities/data-source'
import {
  ImportStatusBadge,
  importStatusLabels,
  useImports,
  type ImportBatch,
  type ImportStatus,
} from '@/entities/import-batch'
import { UploadImportForm } from '@/features/upload-import'
import { formatDateTime } from '@/shared/lib/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select'

const ALL = 'all'

const columns: Column<ImportBatch>[] = [
  { id: 'createdAt', header: 'Создан', cell: (b) => formatDateTime(b.createdAt) },
  { id: 'source', header: 'Источник', cell: (b) => b.dataSourceName ?? b.dataSourceId },
  { id: 'file', header: 'Файл', cell: (b) => b.fileName ?? 'Синхронизация' },
  { id: 'status', header: 'Статус', cell: (b) => <ImportStatusBadge status={b.status} /> },
  { id: 'total', header: 'Строк', cell: (b) => b.totalRows, className: 'text-right' },
  { id: 'rejected', header: 'Отклонено', cell: (b) => b.rejectedRows, className: 'text-right' },
]

const statusItems = [
  { value: ALL, label: 'Все статусы' },
  ...Object.entries(importStatusLabels).map(([value, label]) => ({ value, label })),
]

export function ImportsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const dataSourceId = searchParams.get('source') ?? ALL
  const status = searchParams.get('status') ?? ALL
  const page = Number(searchParams.get('page') ?? 1)

  const sources = useDataSources()
  const imports = useImports({
    dataSourceId: dataSourceId === ALL ? undefined : dataSourceId,
    status: status === ALL ? undefined : (status as ImportStatus),
    page,
  })

  const updateParams = (changes: Record<string, string | null>) =>
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      for (const [key, value] of Object.entries(changes)) {
        if (value === null || value === ALL) next.delete(key)
        else next.set(key, value)
      }
      return next
    })

  const sourceItems = [
    { value: ALL, label: 'Все источники' },
    ...(sources.data ?? []).map((s) => ({ value: s.id, label: s.name })),
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Импорт данных" description="Загрузка файлов и история импортов" />

      <Card>
        <CardHeader>
          <CardTitle>Загрузить файл</CardTitle>
        </CardHeader>
        <CardContent>
          <UploadImportForm
            onUploaded={(batchId, isDuplicate) =>
              navigate(`/import/${batchId}`, { state: { isDuplicate } })
            }
          />
        </CardContent>
      </Card>

      <section className="flex flex-col gap-4">
        <div className="flex flex-wrap gap-2">
          <Select
            items={sourceItems}
            value={dataSourceId}
            onValueChange={(value) => updateParams({ source: value, page: null })}
          >
            <SelectTrigger className="w-56" aria-label="Источник">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {sourceItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select
            items={statusItems}
            value={status}
            onValueChange={(value) => updateParams({ status: value, page: null })}
          >
            <SelectTrigger className="w-56" aria-label="Статус">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {statusItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <DataTable
          columns={columns}
          rows={imports.data?.items}
          getRowId={(b) => b.id}
          isLoading={imports.isLoading}
          error={imports.error}
          emptyTitle="Импортов пока нет"
          meta={imports.data?.meta}
          onPageChange={(next) => updateParams({ page: String(next) })}
          onRowClick={(b) => navigate(`/import/${b.id}`)}
        />
      </section>
    </div>
  )
}
