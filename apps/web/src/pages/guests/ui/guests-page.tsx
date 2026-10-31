import { useMemo } from 'react'
import { useNavigate } from 'react-router'
import {
  GuestStatusBadge,
  MaskedIdentifierList,
  useGuestSearch,
  type GuestListItem,
  type GuestStatus,
} from '@/entities/guest'
import { useAuth } from '@/entities/user'
import { GUEST_FILTER_KEYS, GuestFiltersPanel } from '@/features/filter-guests'
import { formatDateTime, formatMoney } from '@/shared/lib/format'
import { formatSort, useListParams } from '@/shared/lib/list-params'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'

const AMOUNT_PERMISSION = 'field.guest.total_amount.read'
const IDENTIFIERS_PERMISSION = 'field.guest.identifiers.read'
const SORTABLE_FIELDS = ['eventCount', 'totalAmountMinor', 'firstSeenAt', 'lastSeenAt']

const formatDate = (value?: string | null) => (value ? formatDateTime(value) : '—')

const identityColumns: Column<GuestListItem>[] = [
  { id: 'name', header: 'Гость', cell: (g) => g.displayName ?? 'Без имени' },
  { id: 'status', header: 'Статус', cell: (g) => <GuestStatusBadge status={g.status} /> },
]

const identifiersColumn: Column<GuestListItem> = {
  id: 'identifiers',
  header: 'Идентификаторы',
  cell: (g) => <MaskedIdentifierList identifiers={g.maskedIdentifiers} />,
}

const countColumns: Column<GuestListItem>[] = [
  { id: 'tenants', header: 'Арендаторов', cell: (g) => g.tenantCount, className: 'text-right' },
  {
    id: 'eventCount',
    header: 'Событий',
    cell: (g) => g.eventCount ?? 0,
    className: 'text-right',
    sortable: true,
  },
]

const amountColumn: Column<GuestListItem> = {
  id: 'totalAmountMinor',
  header: 'Сумма покупок',
  cell: (g) => (g.totalAmountMinor ? formatMoney(g.totalAmountMinor) : '—'),
  className: 'text-right',
  sortable: true,
}

const activityColumns: Column<GuestListItem>[] = [
  {
    id: 'firstSeenAt',
    header: 'Первое появление',
    cell: (g) => formatDate(g.firstSeenAt),
    sortable: true,
  },
  {
    id: 'lastSeenAt',
    header: 'Последняя активность',
    cell: (g) => formatDate(g.lastSeenAt),
    sortable: true,
  },
]

export function GuestsPage() {
  const navigate = useNavigate()
  const { can } = useAuth()
  const params = useListParams()
  const { get, page, pageSize, setPage, setPageSize, setSort, reset } = params
  const canSeeAmount = can(AMOUNT_PERMISSION)
  const canSeeIdentifiers = can(IDENTIFIERS_PERMISSION)

  const sort = params.sort && SORTABLE_FIELDS.includes(params.sort.field) ? params.sort : undefined
  const hasPurchases = get('has_purchases')
  const isFiltered = GUEST_FILTER_KEYS.some((key) => get(key) !== '')

  const columns = useMemo(
    () => [
      ...identityColumns,
      ...(canSeeIdentifiers ? [identifiersColumn] : []),
      ...countColumns,
      ...(canSeeAmount ? [amountColumn] : []),
      ...activityColumns,
    ],
    [canSeeAmount, canSeeIdentifiers],
  )

  const guests = useGuestSearch({
    query: get('q'),
    sourceId: get('source_id') || undefined,
    tenantId: get('tenant_id') || undefined,
    segmentId: get('segment_id') || undefined,
    status: (get('status') as GuestStatus) || undefined,
    hasPurchases: hasPurchases ? hasPurchases === 'true' : undefined,
    periodFrom: get('period_from') || undefined,
    periodTo: get('period_to') || undefined,
    firstSeenFrom: get('first_seen_from') || undefined,
    lastSeenTo: get('last_seen_to') || undefined,
    sort: formatSort(sort),
    page,
    pageSize,
  })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Гости"
        description="Единая база профилей: поиск по телефону, email, карте или имени"
      />
      <GuestFiltersPanel params={params} isFiltered={isFiltered} onReset={reset} />
      <DataTable
        columns={columns}
        rows={guests.data?.items}
        getRowId={(g) => g.id}
        isLoading={guests.isLoading}
        error={guests.error}
        emptyTitle={isFiltered ? 'Гости по фильтрам не найдены' : 'Гостей пока нет'}
        emptyAction={
          isFiltered ? (
            <Button type="button" variant="outline" onClick={reset}>
              Сбросить фильтры
            </Button>
          ) : undefined
        }
        meta={guests.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        sort={sort}
        onSortChange={setSort}
        onRowClick={(g) => navigate(`/guests/${g.id}`)}
      />
    </div>
  )
}
