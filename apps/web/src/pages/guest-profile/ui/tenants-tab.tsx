import type { GuestProfile, GuestTenantSummary } from '@/entities/guest'
import { useAuth } from '@/entities/user'
import { formatDateTime, formatMoney } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'

export function TenantsTab({ profile }: { profile: GuestProfile }) {
  const { can } = useAuth()

  const columns: Column<GuestTenantSummary>[] = [
    { id: 'name', header: 'Арендатор', cell: (t) => t.tenantName },
    { id: 'category', header: 'Категория', cell: (t) => t.categoryName ?? '—' },
    { id: 'floor', header: 'Этаж', cell: (t) => t.floor ?? '—', className: 'text-right' },
    { id: 'events', header: 'Событий', cell: (t) => t.eventCount, className: 'text-right' },
    {
      id: 'purchases',
      header: 'Покупок',
      cell: (t) => t.purchaseCount ?? 0,
      className: 'text-right',
    },
    ...(can('field.event.amount.read')
      ? [
          {
            id: 'amount',
            header: 'Потрачено',
            cell: (t: GuestTenantSummary) => formatMoney(t.totalAmountMinor ?? 0),
            className: 'text-right',
          },
        ]
      : []),
    {
      id: 'lastVisit',
      header: 'Последний визит',
      cell: (t) => (t.lastVisitAt ? formatDateTime(t.lastVisitAt) : '—'),
    },
  ]

  return (
    <DataTable
      columns={columns}
      rows={profile.tenants}
      getRowId={(t) => t.tenantId}
      emptyTitle="Активности по арендаторам нет"
    />
  )
}
