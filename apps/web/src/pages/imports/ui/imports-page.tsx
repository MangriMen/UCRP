import { useNavigate, useSearchParams } from 'react-router'
import { useDataSources } from '@/entities/data-source'
import {
  ImportStatusBadge,
  importStatusLabels,
  useImports,
  type ImportBatch,
  type ImportStatus,
} from '@/entities/import-batch'
import { useAuth } from '@/entities/user'
import { UploadImportForm } from '@/features/upload-import'
import { formatDateTime } from '@/shared/lib/format'
import { ALL_OPTION, parsePage } from '@/shared/lib/search-params'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'

const columns: Column<ImportBatch>[] = [
  { id: 'createdAt', header: 'Создан', cell: (b) => formatDateTime(b.createdAt) },
  { id: 'source', header: 'Источник', cell: (b) => b.dataSourceName ?? b.dataSourceId },
  { id: 'file', header: 'Файл', cell: (b) => b.fileName ?? 'Синхронизация' },
  { id: 'status', header: 'Статус', cell: (b) => <ImportStatusBadge status={b.status} /> },
  { id: 'total', header: 'Строк', cell: (b) => b.totalRows, className: 'text-right' },
  { id: 'rejected', header: 'Отклонено', cell: (b) => b.rejectedRows, className: 'text-right' },
]

const statusItems = [
  { value: ALL_OPTION, label: 'Все статусы' },
  ...Object.entries(importStatusLabels).map(([value, label]) => ({ value, label })),
]

const isImportStatus = (value: string): value is ImportStatus => value in importStatusLabels

export function ImportsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const dataSourceId = searchParams.get('source') ?? ALL_OPTION
  const status = searchParams.get('status') ?? ALL_OPTION
  const page = parsePage(searchParams.get('page'))

  const { can } = useAuth()
  const sources = useDataSources()
  const imports = useImports({
    dataSourceId: dataSourceId === ALL_OPTION ? undefined : dataSourceId,
    status: isImportStatus(status) ? status : undefined,
    page,
  })

  const updateParams = (changes: Record<string, string | null>) =>
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      for (const [key, value] of Object.entries(changes)) {
        if (value === null || value === ALL_OPTION) next.delete(key)
        else next.set(key, value)
      }
      return next
    })

  const sourceItems = [
    { value: ALL_OPTION, label: 'Все источники' },
    ...(sources.data ?? []).map((s) => ({ value: s.id, label: s.name })),
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Импорт данных" description="Загрузка файлов и история импортов" />

      {can('import.create') && (
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
      )}

      <section className="flex flex-col gap-4">
        <div className="flex flex-wrap gap-2">
          <OptionSelect
            items={sourceItems}
            value={dataSourceId}
            onValueChange={(value) => updateParams({ source: value, page: null })}
            ariaLabel="Источник"
            className="w-56"
          />
          <OptionSelect
            items={statusItems}
            value={status}
            onValueChange={(value) => updateParams({ status: value, page: null })}
            ariaLabel="Статус"
            className="w-56"
          />
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
