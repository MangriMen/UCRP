import { useSegmentMembers, type SegmentMember } from '@/entities/segment'
import { formatDateTime } from '@/shared/lib/format'
import { useListParams } from '@/shared/lib/list-params'
import { DataTable, type Column } from '@/shared/ui/data-table'

const columns: Column<SegmentMember>[] = [
  { id: 'guest', header: 'Гость', cell: (m) => m.displayName ?? 'Без имени' },
  {
    id: 'tenants',
    header: 'Арендаторов',
    cell: (m) => m.tenantCount ?? 0,
    className: 'text-right',
  },
  {
    id: 'lastSeen',
    header: 'Последняя активность',
    cell: (m) => (m.lastSeenAt ? formatDateTime(m.lastSeenAt) : '—'),
  },
]

export function MemberTable({ segmentId }: { segmentId: string }) {
  const { page, setPage } = useListParams()
  const members = useSegmentMembers(segmentId, { page })

  return (
    <DataTable
      columns={columns}
      rows={members.data?.items}
      getRowId={(m) => m.guestId}
      isLoading={members.isLoading}
      error={members.error}
      emptyTitle="В сегменте пока нет гостей"
      meta={members.data?.meta}
      onPageChange={setPage}
    />
  )
}
