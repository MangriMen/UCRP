import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
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
import { useDebouncedValue } from '@/shared/lib/hooks/use-debounced-value'
import { ALL_OPTION, parsePage } from '@/shared/lib/search-params'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'

const SEARCH_DEBOUNCE_MS = 300

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

const isSegmentStatus = (value: string): value is SegmentStatus => value in segmentStatusLabels

export function SegmentsPage() {
  const navigate = useNavigate()
  const { can } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()
  const urlQuery = searchParams.get('q') ?? ''
  const status = searchParams.get('status') ?? ALL_OPTION
  const page = parsePage(searchParams.get('page'))

  const [input, setInput] = useState(urlQuery)
  const query = useDebouncedValue(input.trim(), SEARCH_DEBOUNCE_MS)

  const updateParams = useCallback(
    (changes: Record<string, string | null>) =>
      setSearchParams((current) => {
        const next = new URLSearchParams(current)
        for (const [key, value] of Object.entries(changes)) {
          if (value === null || value === ALL_OPTION || value === '') next.delete(key)
          else next.set(key, value)
        }
        return next
      }),
    [setSearchParams],
  )

  useEffect(() => {
    if (query !== urlQuery) updateParams({ q: query, page: null })
  }, [query, urlQuery, updateParams])

  const segments = useSegments({
    query: urlQuery,
    status: isSegmentStatus(status) ? status : undefined,
    page,
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
          value={input}
          onChange={(event) => setInput(event.target.value)}
          placeholder="Название сегмента"
          aria-label="Поиск сегментов"
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
        rows={segments.data?.items}
        getRowId={(s) => s.id}
        isLoading={segments.isLoading}
        error={segments.error}
        emptyTitle="Сегментов пока нет"
        meta={segments.data?.meta}
        onPageChange={(next) => updateParams({ page: String(next) })}
        onRowClick={(s) => navigate(`/segments/${s.id}`)}
      />
    </div>
  )
}
