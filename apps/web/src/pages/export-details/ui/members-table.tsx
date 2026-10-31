import { useNavigate } from 'react-router'
import { identifierTypeLabels } from '@/entities/guest'
import { useExportMembers, type ExportMember } from '@/entities/export'
import { useAuth } from '@/entities/user'
import { useListParams } from '@/shared/lib/list-params'
import { DataTable, type Column } from '@/shared/ui/data-table'

const columns: Column<ExportMember>[] = [
  { id: 'guest', header: 'Гость', cell: (m) => m.displayName ?? m.guestId },
  {
    id: 'contactType',
    header: 'Тип контакта',
    cell: (m) => (m.contact ? identifierTypeLabels[m.contact.idType] : '—'),
  },
  { id: 'contact', header: 'Контакт', cell: (m) => m.contact?.value ?? '—' },
]

interface MembersTableProps {
  exportId: string
  isControl: boolean
}

export function MembersTable({ exportId, isControl }: MembersTableProps) {
  const navigate = useNavigate()
  const { can } = useAuth()
  const { page, pageSize, setPage, setPageSize } = useListParams()
  const members = useExportMembers(exportId, { isControl, page, pageSize })

  return (
    <DataTable
      columns={columns}
      rows={members.data?.items}
      getRowId={(m) => m.guestId}
      isLoading={members.isLoading}
      error={members.error}
      emptyTitle={isControl ? 'Контрольная группа не выделялась' : 'Получателей нет'}
      meta={members.data?.meta}
      onPageChange={setPage}
      onPageSizeChange={setPageSize}
      onRowClick={can('guest.profile.read') ? (m) => navigate(`/guests/${m.guestId}`) : undefined}
    />
  )
}
