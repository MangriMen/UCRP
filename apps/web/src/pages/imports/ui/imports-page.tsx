import { useNavigate } from 'react-router'
import { useDataSourceOptions } from '@/entities/data-source'
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
import { useOptionParam } from '@/shared/lib/hooks/use-option-param'
import { useListParams } from '@/shared/lib/list-params'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'

const columns: Column<ImportBatch>[] = [
  { id: 'createdAt', header: 'Создана', cell: (b) => formatDateTime(b.createdAt) },
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

const statusValues = statusItems.map((item) => item.value)

export function ImportsPage() {
  const navigate = useNavigate()
  const params = useListParams()
  const { page, pageSize, setPage, setPageSize } = params

  const { can } = useAuth()
  const sources = useDataSourceOptions()
  const sourceParam = useOptionParam(
    params,
    'source',
    sources.isSuccess ? sources.data.map((s) => s.id) : undefined,
  )
  const statusParam = useOptionParam(params, 'status', statusValues)
  const imports = useImports({
    dataSourceId: sourceParam.value === ALL_OPTION ? undefined : sourceParam.value,
    status: statusParam.value === ALL_OPTION ? undefined : (statusParam.value as ImportStatus),
    page,
    pageSize,
  })

  const sourceItems = [
    { value: ALL_OPTION, label: 'Все источники' },
    ...(sources.data ?? []).map((s) => ({ value: s.id, label: s.name })),
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Импорт данных" description="Загрузка файлов и журнал загрузок" />

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
            value={sourceParam.value}
            onValueChange={sourceParam.onChange}
            ariaLabel="Источник"
            className="w-56"
          />
          <OptionSelect
            items={statusItems}
            value={statusParam.value}
            onValueChange={statusParam.onChange}
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
          emptyTitle="Загрузок пока нет"
          meta={imports.data?.meta}
          onPageChange={setPage}
          onPageSizeChange={setPageSize}
          onRowClick={(b) => navigate(`/import/${b.id}`)}
        />
      </section>
    </div>
  )
}
