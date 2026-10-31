import { useNavigate, useSearchParams } from 'react-router'
import {
  MergeResolutionBadge,
  mergeResolutionLabels,
  useMergeCandidates,
  type MergeCandidate,
  type MergeResolution,
} from '@/entities/merge-candidate'
import { formatDateTime } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select'

const ALL = 'all'
const DEFAULT_RESOLUTION: MergeResolution = 'pending'

const columns: Column<MergeCandidate>[] = [
  { id: 'createdAt', header: 'Создана', cell: (c) => formatDateTime(c.createdAt) },
  { id: 'guestA', header: 'Профиль A', cell: (c) => c.guestAId },
  { id: 'guestB', header: 'Профиль B', cell: (c) => c.guestBId },
  { id: 'reason', header: 'Причина', cell: (c) => c.reasonText ?? c.reasonCode },
  { id: 'score', header: 'Уверенность', cell: (c) => `${c.score}%`, className: 'text-right' },
  {
    id: 'resolution',
    header: 'Статус',
    cell: (c) => <MergeResolutionBadge resolution={c.resolution} />,
  },
]

const resolutionItems = [
  { value: ALL, label: 'Все статусы' },
  ...Object.entries(mergeResolutionLabels).map(([value, label]) => ({ value, label })),
]

const isResolution = (value: string): value is MergeResolution => value in mergeResolutionLabels

const parsePage = (value: string | null) => {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

export function MergeQueuePage() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const resolution = searchParams.get('resolution') ?? DEFAULT_RESOLUTION
  const page = parsePage(searchParams.get('page'))

  const candidates = useMergeCandidates({
    resolution: isResolution(resolution) ? resolution : undefined,
    page,
  })

  const updateParams = (changes: Record<string, string | null>) =>
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      for (const [key, value] of Object.entries(changes)) {
        if (value === null) next.delete(key)
        else next.set(key, value)
      }
      return next
    })

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Склейка профилей"
        description="Спорные пары профилей и история решений"
        actions={
          <Select
            items={resolutionItems}
            value={resolution}
            onValueChange={(value) => updateParams({ resolution: value, page: null })}
          >
            <SelectTrigger className="w-56" aria-label="Статус">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {resolutionItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        }
      />
      <DataTable
        columns={columns}
        rows={candidates.data?.items}
        getRowId={(c) => c.id}
        isLoading={candidates.isLoading}
        error={candidates.error}
        emptyTitle="Заявок нет"
        meta={candidates.data?.meta}
        onPageChange={(next) => updateParams({ page: String(next) })}
        onRowClick={(c) => navigate(`/identity/queue/${c.id}`)}
      />
    </div>
  )
}
