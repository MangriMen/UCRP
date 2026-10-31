import { useMemo } from 'react'
import { Link } from 'react-router'
import { useTenants, type TenantListItem } from '@/entities/tenant'
import { formatAggregateCell } from '@/entities/tenant-aggregates'
import { useAuth } from '@/entities/user'
import type { Period } from '@/shared/lib/period'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'

const VISITS_PERMISSION = 'metric.visits_count.read'
const NEW_GUESTS_PERMISSION = 'metric.new_guests_via_mall.read'
const SUPPRESSION_THRESHOLD = 5
const TOP_SIZE = 10

const statCell = (value: number | null | undefined) =>
  formatAggregateCell(value, SUPPRESSION_THRESHOLD)

const baseColumns: Column<TenantListItem>[] = [
  {
    id: 'name',
    header: 'Арендатор',
    cell: (t) => (
      <Link to={`/tenants/${t.id}`} className="font-medium hover:underline">
        {t.name}
      </Link>
    ),
  },
  { id: 'category', header: 'Категория', cell: (t) => t.categoryName },
]

const visitColumns: Column<TenantListItem>[] = [
  {
    id: 'guestsCount',
    header: 'Гостей',
    cell: (t) => statCell(t.guestsCount),
    className: 'text-right',
  },
  {
    id: 'visitsCount',
    header: 'Визитов',
    cell: (t) => statCell(t.visitsCount),
    className: 'text-right',
  },
]

const newGuestsColumn: Column<TenantListItem> = {
  id: 'newGuestsCount',
  header: 'Новых гостей',
  cell: (t) => statCell(t.newGuestsCount),
  className: 'text-right',
}

export function TenantBreakdown({ period }: { period: Period }) {
  const { can } = useAuth()
  const canSeeVisits = can(VISITS_PERMISSION)
  const canSeeNewGuests = can(NEW_GUESTS_PERMISSION)

  const columns = useMemo(
    () => [
      ...baseColumns,
      ...(canSeeVisits ? visitColumns : []),
      ...(canSeeNewGuests ? [newGuestsColumn] : []),
    ],
    [canSeeVisits, canSeeNewGuests],
  )

  const tenants = useTenants({
    periodFrom: period.from,
    periodTo: period.to,
    sort: canSeeVisits ? '-visitsCount' : undefined,
    pageSize: TOP_SIZE,
  })

  const allTenantsQuery = new URLSearchParams({ period_from: period.from, period_to: period.to })

  return (
    <Card>
      <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-2">
        <CardTitle>Арендаторы за период</CardTitle>
        <Button variant="outline" size="sm" render={<Link to={`/tenants?${allTenantsQuery}`} />}>
          Все арендаторы
        </Button>
      </CardHeader>
      <CardContent>
        <DataTable
          columns={columns}
          rows={tenants.data?.items}
          getRowId={(t) => t.id}
          isLoading={tenants.isLoading}
          error={tenants.error}
          emptyTitle="За период арендаторов нет"
        />
      </CardContent>
    </Card>
  )
}
