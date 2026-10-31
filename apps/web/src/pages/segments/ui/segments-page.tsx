import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router'
import {
  SegmentStatusBadge,
  segmentStatusLabels,
  useSegments,
  type Segment,
  type SegmentStatus,
} from '@/entities/segment'
import { formatDateTime } from '@/shared/lib/format'
import { useDebouncedValue } from '@/shared/lib/hooks/use-debounced-value'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { Input } from '@/shared/ui/input'
import { PageHeader } from '@/shared/ui/page-header'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select'

const ALL = 'all'
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
  { value: ALL, label: 'Все статусы' },
  ...Object.entries(segmentStatusLabels).map(([value, label]) => ({ value, label })),
]

const isSegmentStatus = (value: string): value is SegmentStatus => value in segmentStatusLabels

const parsePage = (value: string | null) => {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

export function SegmentsPage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const urlQuery = searchParams.get('q') ?? ''
  const status = searchParams.get('status') ?? ALL
  const page = parsePage(searchParams.get('page'))

  const [input, setInput] = useState(urlQuery)
  const query = useDebouncedValue(input.trim(), SEARCH_DEBOUNCE_MS)

  const updateParams = useCallback(
    (changes: Record<string, string | null>) =>
      setSearchParams((current) => {
        const next = new URLSearchParams(current)
        for (const [key, value] of Object.entries(changes)) {
          if (value === null || value === ALL || value === '') next.delete(key)
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
      <PageHeader title="Сегменты" description="Группы гостей и размер аудитории" />
      <div className="flex flex-wrap gap-2">
        <Input
          value={input}
          onChange={(event) => setInput(event.target.value)}
          placeholder="Название сегмента"
          aria-label="Поиск сегментов"
          className="w-56"
        />
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
