import { SegmentStatusBadge, useSegmentVersions, type SegmentVersion } from '@/entities/segment'
import { formatDateTime } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'

const columns: Column<SegmentVersion>[] = [
  { id: 'version', header: 'Версия', cell: (v) => `v${v.versionNo}`, className: 'w-24' },
  { id: 'rule', header: 'Правило', cell: (v) => v.ruleText ?? '—' },
  { id: 'status', header: 'Статус', cell: (v) => <SegmentStatusBadge status={v.status} /> },
  { id: 'members', header: 'Аудитория', cell: (v) => v.memberCount, className: 'text-right' },
  { id: 'createdAt', header: 'Создана', cell: (v) => formatDateTime(v.createdAt ?? '') },
]

export function VersionHistory({ segmentId }: { segmentId: string }) {
  const versions = useSegmentVersions(segmentId)

  return (
    <DataTable
      columns={columns}
      rows={versions.data}
      getRowId={(v) => v.id}
      isLoading={versions.isLoading}
      error={versions.error}
      emptyTitle="Версий нет"
    />
  )
}
