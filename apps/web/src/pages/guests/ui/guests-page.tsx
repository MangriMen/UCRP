import { useNavigate } from 'react-router'
import { GuestStatusBadge, useGuestSearch, type Guest } from '@/entities/guest'
import { GuestFiltersPanel, useGuestFilters } from '@/features/filter-guests'
import { formatDateTime } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'

const columns: Column<Guest>[] = [
  { id: 'name', header: 'Гость', cell: (g) => g.displayName ?? 'Без имени' },
  { id: 'status', header: 'Статус', cell: (g) => <GuestStatusBadge status={g.status} /> },
  { id: 'tenants', header: 'Арендаторов', cell: (g) => g.tenantCount, className: 'text-right' },
  { id: 'events', header: 'Событий', cell: (g) => g.eventCount ?? 0, className: 'text-right' },
  {
    id: 'lastSeen',
    header: 'Последняя активность',
    cell: (g) => (g.lastSeenAt ? formatDateTime(g.lastSeenAt) : '—'),
  },
]

export function GuestsPage() {
  const navigate = useNavigate()
  const { filters, update, setPage, reset, isFiltered } = useGuestFilters()

  const guests = useGuestSearch({
    query: filters.q,
    sourceId: filters.sourceId || undefined,
    tenantId: filters.tenantId || undefined,
    periodFrom: filters.periodFrom || undefined,
    periodTo: filters.periodTo || undefined,
    page: filters.page,
  })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Гости"
        description="Единая база профилей: поиск по телефону, email, карте или имени"
      />
      <GuestFiltersPanel
        filters={filters}
        isFiltered={isFiltered}
        onChange={update}
        onReset={reset}
      />
      <DataTable
        columns={columns}
        rows={guests.data?.items}
        getRowId={(g) => g.id}
        isLoading={guests.isLoading}
        error={guests.error}
        emptyTitle="Гости не найдены"
        meta={guests.data?.meta}
        onPageChange={setPage}
        onRowClick={(g) => navigate(`/guests/${g.id}`)}
      />
    </div>
  )
}
