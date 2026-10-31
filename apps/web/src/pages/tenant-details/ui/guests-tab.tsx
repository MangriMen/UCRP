import { useMemo } from 'react'
import { useNavigate } from 'react-router'
import { GuestStatusBadge, useTenantGuests, type Guest } from '@/entities/guest'
import { useAuth } from '@/entities/user'
import { formatDateTime, formatMoney } from '@/shared/lib/format'
import { formatSort, useListParams } from '@/shared/lib/list-params'
import { DataTable, type Column } from '@/shared/ui/data-table'

const AMOUNT_PERMISSION = 'field.guest.total_amount.read'

const formatDate = (value?: string | null) => (value ? formatDateTime(value) : '—')

const baseColumns: Column<Guest>[] = [
  { id: 'name', header: 'Гость', cell: (g) => g.displayName ?? 'Без имени' },
  { id: 'status', header: 'Статус', cell: (g) => <GuestStatusBadge status={g.status} /> },
  {
    id: 'eventCount',
    header: 'Событий',
    cell: (g) => g.eventCount ?? 0,
    className: 'text-right',
    sortable: true,
  },
]

const amountColumn: Column<Guest> = {
  id: 'totalAmountMinor',
  header: 'Сумма покупок',
  cell: (g) => (g.totalAmountMinor ? formatMoney(g.totalAmountMinor) : '—'),
  className: 'text-right',
  sortable: true,
}

const activityColumn: Column<Guest> = {
  id: 'lastSeenAt',
  header: 'Последняя активность',
  cell: (g) => formatDate(g.lastSeenAt),
  sortable: true,
}

export function GuestsTab({ tenantId }: { tenantId: string }) {
  const navigate = useNavigate()
  const { can } = useAuth()
  const params = useListParams()
  const { page, pageSize, setPage, setPageSize, setSort } = params
  const canSeeAmount = can(AMOUNT_PERMISSION)

  const columns = useMemo(
    () => [...baseColumns, ...(canSeeAmount ? [amountColumn] : []), activityColumn],
    [canSeeAmount],
  )
  const sortable = columns.filter((column) => column.sortable).map((column) => column.id)
  const sort = params.sort && sortable.includes(params.sort.field) ? params.sort : undefined

  const guests = useTenantGuests(tenantId, { sort: formatSort(sort), page, pageSize })

  return (
    <DataTable
      columns={columns}
      rows={guests.data?.items}
      getRowId={(g) => g.id}
      isLoading={guests.isLoading}
      error={guests.error}
      emptyTitle="У арендатора пока нет гостей"
      meta={guests.data?.meta}
      onPageChange={setPage}
      onPageSizeChange={setPageSize}
      sort={sort}
      onSortChange={setSort}
      onRowClick={(g) => navigate(`/guests/${g.id}`)}
    />
  )
}
