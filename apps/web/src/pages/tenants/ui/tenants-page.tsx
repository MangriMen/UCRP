import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router'
import { useTenants, type TenantListItem } from '@/entities/tenant'
import { formatAggregateCell } from '@/entities/tenant-aggregates'
import { useAuth } from '@/entities/user'
import { TenantFormDialog } from '@/features/manage-tenants'
import { formatSort, useListParams } from '@/shared/lib/list-params'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'
import { StatusBadge } from '@/shared/ui/status-badge'
import { TENANT_FILTER_KEYS } from '../model/filter-keys'
import { TenantFilters } from './tenant-filters'

const VISITS_PERMISSION = 'metric.visits_count.read'
const NEW_GUESTS_PERMISSION = 'metric.new_guests_via_mall.read'
const SUPPRESSION_THRESHOLD = 5

const statCell = (value: number | null | undefined) =>
  formatAggregateCell(value, SUPPRESSION_THRESHOLD)

const sourcesCell = (tenant: TenantListItem) => (
  <div className="flex flex-wrap items-center gap-2">
    <span>{tenant.sourcesCount}</span>
    {tenant.problemSourcesCount > 0 && (
      <StatusBadge tone="danger">{`Проблемы: ${tenant.problemSourcesCount}`}</StatusBadge>
    )}
  </div>
)

const baseColumns: Column<TenantListItem>[] = [
  {
    id: 'name',
    header: 'Арендатор',
    cell: (t) => (
      <div className="flex flex-col">
        <span className="font-medium">{t.name}</span>
        <span className="text-xs text-muted-foreground">{t.code}</span>
      </div>
    ),
  },
  { id: 'category', header: 'Категория', cell: (t) => t.categoryName },
  {
    id: 'floor',
    header: 'Этаж',
    cell: (t) => t.floor ?? '—',
    className: 'text-right',
    sortable: true,
  },
]

const visitColumns: Column<TenantListItem>[] = [
  {
    id: 'guestsCount',
    header: 'Гостей',
    cell: (t) => statCell(t.guestsCount),
    className: 'text-right',
    sortable: true,
  },
  {
    id: 'visitsCount',
    header: 'Визитов',
    cell: (t) => statCell(t.visitsCount),
    className: 'text-right',
    sortable: true,
  },
]

const newGuestsColumn: Column<TenantListItem> = {
  id: 'newGuestsCount',
  header: 'Новых гостей',
  cell: (t) => statCell(t.newGuestsCount),
  className: 'text-right',
  sortable: true,
}

const tailColumns: Column<TenantListItem>[] = [
  { id: 'status', header: 'Статус', cell: (t) => <ActiveBadge isActive={t.isActive} /> },
  { id: 'sources', header: 'Источники данных', cell: sourcesCell },
]

export function TenantsPage() {
  const navigate = useNavigate()
  const { can } = useAuth()
  const params = useListParams()
  const { get, page, pageSize, setPage, setPageSize, setSort, reset } = params
  const [creating, setCreating] = useState(false)
  const canSeeVisits = can(VISITS_PERMISSION)
  const canSeeNewGuests = can(NEW_GUESTS_PERMISSION)

  const columns = useMemo(
    () => [
      ...baseColumns,
      ...(canSeeVisits ? visitColumns : []),
      ...(canSeeNewGuests ? [newGuestsColumn] : []),
      ...tailColumns,
    ],
    [canSeeVisits, canSeeNewGuests],
  )

  const sortableFields = useMemo(
    () => columns.filter((column) => column.sortable).map((column) => column.id),
    [columns],
  )
  const sort = params.sort && sortableFields.includes(params.sort.field) ? params.sort : undefined
  const isActive = get('is_active')
  const floor = get('floor')
  const isFiltered = TENANT_FILTER_KEYS.some((key) => get(key) !== '')

  const tenants = useTenants({
    q: get('q') || undefined,
    categoryId: get('category_id') || undefined,
    floor: floor === '' ? undefined : Number(floor),
    isActive: isActive ? isActive === 'true' : undefined,
    periodFrom: get('period_from') || undefined,
    periodTo: get('period_to') || undefined,
    sort: formatSort(sort),
    page,
    pageSize,
  })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Арендаторы"
        description="Статистика по арендаторам за период, статус и источники данных"
        actions={
          can('reference.write') && (
            <Button onClick={() => setCreating(true)}>Добавить арендатора</Button>
          )
        }
      />
      <TenantFilters params={params} isFiltered={isFiltered} onReset={reset} />
      <DataTable
        columns={columns}
        rows={tenants.data?.items}
        getRowId={(t) => t.id}
        isLoading={tenants.isLoading}
        error={tenants.error}
        emptyTitle={isFiltered ? 'Арендаторы по фильтрам не найдены' : 'Арендаторов пока нет'}
        emptyAction={
          isFiltered ? (
            <Button type="button" variant="outline" onClick={reset}>
              Сбросить фильтры
            </Button>
          ) : undefined
        }
        meta={tenants.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        sort={sort}
        onSortChange={setSort}
        onRowClick={(t) => navigate(`/tenants/${t.id}`)}
      />
      {creating && (
        <TenantFormDialog open tenant={null} onOpenChange={(open) => !open && setCreating(false)} />
      )}
    </div>
  )
}
