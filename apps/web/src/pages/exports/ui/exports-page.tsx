import { useNavigate } from 'react-router'
import {
  DownloadExportButton,
  ExportStatusBadge,
  exportChannelLabels,
  exportStatusLabels,
  useExports,
  type Export,
  type ExportChannel,
  type ExportStatus,
} from '@/entities/export'
import { useAuth } from '@/entities/user'
import { CreateExportButton } from '@/features/create-export'
import { formatDateTime, formatNumber } from '@/shared/lib/format'
import { useListParams } from '@/shared/lib/list-params'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { OptionSelect, type Option } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'

const FILTER_KEYS = ['channel', 'status', 'segment_id', 'campaign_id']

const channelOptions: Option[] = [
  { value: ALL_OPTION, label: 'Все каналы' },
  ...Object.entries(exportChannelLabels).map(([value, label]) => ({ value, label })),
]

const statusOptions: Option[] = [
  { value: ALL_OPTION, label: 'Все статусы' },
  ...Object.entries(exportStatusLabels).map(([value, label]) => ({ value, label })),
]

const downloadColumn: Column<Export> = {
  id: 'download',
  header: '',
  className: 'text-right',
  cell: (e) =>
    e.status === 'ready' ? <DownloadExportButton exportId={e.id} format={e.format} /> : null,
}

const baseColumns: Column<Export>[] = [
  { id: 'createdAt', header: 'Создана', cell: (e) => formatDateTime(e.createdAt) },
  {
    id: 'segment',
    header: 'Сегмент',
    cell: (e) =>
      `${e.segmentName ?? e.segmentId}${e.segmentVersionNo != null ? ` · v${e.segmentVersionNo}` : ''}`,
  },
  { id: 'channel', header: 'Канал', cell: (e) => exportChannelLabels[e.channel] },
  { id: 'status', header: 'Статус', cell: (e) => <ExportStatusBadge status={e.status} /> },
  {
    id: 'size',
    header: 'Получателей',
    cell: (e) => (e.exportedCount == null ? 'скрыто' : formatNumber(e.exportedCount)),
    className: 'text-right',
  },
]

export function ExportsPage() {
  const navigate = useNavigate()
  const { can } = useAuth()
  const { get, update, page, pageSize, setPage, setPageSize, reset } = useListParams()
  const columns = can('export.download') ? [...baseColumns, downloadColumn] : baseColumns
  const isFiltered = FILTER_KEYS.some((key) => get(key) !== '')
  const exports = useExports({
    channel: (get('channel') as ExportChannel) || undefined,
    status: (get('status') as ExportStatus) || undefined,
    segmentId: get('segment_id') || undefined,
    campaignId: get('campaign_id') || undefined,
    page,
    pageSize,
  })

  const selectValue = (key: string) => get(key) || ALL_OPTION
  const changeSelect = (key: string) => (value: string) =>
    update({ [key]: value === ALL_OPTION ? null : value })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Выгрузки"
        description="История выгрузок сегментов в каналы коммуникаций"
        actions={can('export.create') ? <CreateExportButton /> : undefined}
      />
      <div className="flex flex-wrap items-center gap-3">
        <OptionSelect
          ariaLabel="Канал"
          className="w-44"
          items={channelOptions}
          value={selectValue('channel')}
          onValueChange={changeSelect('channel')}
        />
        <OptionSelect
          ariaLabel="Статус"
          className="w-44"
          items={statusOptions}
          value={selectValue('status')}
          onValueChange={changeSelect('status')}
        />
        {isFiltered && (
          <Button type="button" variant="ghost" onClick={reset}>
            Сбросить
          </Button>
        )}
      </div>
      <DataTable
        columns={columns}
        rows={exports.data?.items}
        getRowId={(e) => e.id}
        isLoading={exports.isLoading}
        error={exports.error}
        emptyTitle={isFiltered ? 'Выгрузки по фильтрам не найдены' : 'Выгрузок пока нет'}
        emptyAction={
          isFiltered ? (
            <Button type="button" variant="outline" onClick={reset}>
              Сбросить фильтры
            </Button>
          ) : undefined
        }
        meta={exports.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        onRowClick={(e) => navigate(`/exports/${e.id}`)}
      />
    </div>
  )
}
