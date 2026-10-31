import { useSearchParams } from 'react-router'
import {
  DownloadExportButton,
  ExportStatusBadge,
  exportChannelLabels,
  useExports,
  type Export,
} from '@/entities/export'
import { useAuth } from '@/entities/user'
import { CreateExportButton } from '@/features/create-export'
import { formatDateTime } from '@/shared/lib/format'
import { parsePage } from '@/shared/lib/search-params'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'

const downloadColumn: Column<Export> = {
  id: 'download',
  header: '',
  className: 'text-right',
  cell: (e) =>
    e.status === 'ready' ? <DownloadExportButton exportId={e.id} format={e.format} /> : null,
}

const baseColumns: Column<Export>[] = [
  { id: 'createdAt', header: 'Создана', cell: (e) => formatDateTime(e.createdAt) },
  { id: 'segment', header: 'Сегмент', cell: (e) => e.segmentName ?? e.segmentId },
  { id: 'channel', header: 'Канал', cell: (e) => exportChannelLabels[e.channel] },
  { id: 'status', header: 'Статус', cell: (e) => <ExportStatusBadge status={e.status} /> },
  { id: 'size', header: 'Получателей', cell: (e) => e.exportedCount, className: 'text-right' },
]

export function ExportsPage() {
  const { can } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const columns = can('export.download') ? [...baseColumns, downloadColumn] : baseColumns
  const exports = useExports({ page: parsePage(searchParams.get('page')) })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Выгрузки"
        description="История выгрузок сегментов в каналы коммуникаций"
        actions={can('export.create') ? <CreateExportButton /> : undefined}
      />
      <DataTable
        columns={columns}
        rows={exports.data?.items}
        getRowId={(e) => e.id}
        isLoading={exports.isLoading}
        error={exports.error}
        emptyTitle="Выгрузок пока нет"
        meta={exports.data?.meta}
        onPageChange={(page) => setSearchParams({ page: String(page) })}
      />
    </div>
  )
}
