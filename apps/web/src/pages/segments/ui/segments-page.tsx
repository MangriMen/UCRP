import { useNavigate } from 'react-router'
import {
  SegmentStatusBadge,
  segmentStatusLabels,
  useSegments,
  type Segment,
  type SegmentStatus,
} from '@/entities/segment'
import { useAuth } from '@/entities/user'
import { BuildSegmentButton } from '@/features/build-segment'
import { formatDateTime } from '@/shared/lib/format'
import { useDebouncedSearchParam } from '@/shared/lib/hooks/use-debounced-search-param'
import { useOptionParam } from '@/shared/lib/hooks/use-option-param'
import { useListParams } from '@/shared/lib/list-params'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'

const columns: Column<Segment>[] = [
  { id: 'name', header: 'Сегмент', cell: (s) => s.name },
  { id: 'status', header: 'Статус', cell: (s) => <SegmentStatusBadge status={s.status} /> },
  { id: 'members', header: 'Аудитория', cell: (s) => s.memberCount, className: 'text-right' },
  {
    id: 'computedAt',
    header: 'Рассчитан',
    cell: (s) => (s.computedAt ? formatDateTime(s.computedAt) : '—'),
  },
]

const statusItems = [
  { value: ALL_OPTION, label: 'Все статусы' },
  ...Object.entries(segmentStatusLabels).map(([value, label]) => ({ value, label })),
]

const statusValues = statusItems.map((item) => item.value)

export function SegmentsPage() {
  const navigate = useNavigate()
  const { can } = useAuth()
  const params = useListParams()
  const { get, page, pageSize, setPage, setPageSize } = params
  const search = useDebouncedSearchParam(params, 'q')
  const statusParam = useOptionParam(params, 'status', statusValues)

  const segments = useSegments({
    query: get('q'),
    status: statusParam.value === ALL_OPTION ? undefined : (statusParam.value as SegmentStatus),
    page,
    pageSize,
  })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Сегменты"
        description="Группы гостей и размер аудитории"
        actions={can('segment.write') && <BuildSegmentButton />}
      />
      <div className="flex flex-wrap gap-2">
        <Input
          value={search.value}
          onChange={(event) => search.onChange(event.target.value)}
          placeholder="Название сегмента"
          aria-label="Поиск сегментов"
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
        rows={segments.data?.items}
        getRowId={(s) => s.id}
        isLoading={segments.isLoading}
        error={segments.error}
        emptyTitle="Сегментов пока нет"
        meta={segments.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        onRowClick={(s) => navigate(`/segments/${s.id}`)}
      />
    </div>
  )
}
