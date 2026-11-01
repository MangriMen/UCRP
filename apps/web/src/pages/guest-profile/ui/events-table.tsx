import { useEventTypes } from '@/entities/event-type'
import { useGuestEvents, type GuestEvent, type GuestEventSort } from '@/entities/guest'
import { useAuth } from '@/entities/user'
import { hasErrorCode } from '@/shared/api/errors'
import { formatDateTime, formatMoney } from '@/shared/lib/format'
import { useListParams, type SortState } from '@/shared/lib/list-params'
import { isPeriodReversed, PERIOD_REVERSED_MESSAGE } from '@/shared/lib/period'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { OptionSelect } from '@/shared/ui/option-select'
import { PeriodFilter } from '@/shared/ui/period-filter'

const PURCHASE_TYPE = 'purchase'
const DEFAULT_SORT: GuestEventSort = '-occurredAt'
const SORT_FIELDS: Record<string, 'occurredAt' | 'amountMinor'> = {
  occurredAt: 'occurredAt',
  amountMinor: 'amountMinor',
}

interface TenantOption {
  tenantId: string
  tenantName: string
}

interface EventsTableProps {
  guestId: string
  tenants: TenantOption[]
  purchasesOnly?: boolean
}

export function EventsTable({ guestId, tenants, purchasesOnly = false }: EventsTableProps) {
  const { can } = useAuth()
  const params = useListParams()
  const eventTypes = useEventTypes()
  const canSeeAmount = can('field.event.amount.read')

  const type = purchasesOnly ? PURCHASE_TYPE : params.get('type')
  const tenantId = params.get('tenant_id')
  const from = params.get('from')
  const to = params.get('to')
  const isFiltered = Boolean(tenantId || from || to || (!purchasesOnly && type))
  const invalidPeriod = isPeriodReversed({ from, to })

  const sortField = params.sort && SORT_FIELDS[params.sort.field]
  const sort = sortField
    ? (`${params.sort?.direction === 'desc' ? '-' : ''}${sortField}` as GuestEventSort)
    : DEFAULT_SORT

  const shownSort: SortState = params.sort ?? { field: 'occurredAt', direction: 'desc' }

  const events = useGuestEvents(guestId, {
    type: type || undefined,
    tenantId: tenantId || undefined,
    from: from || undefined,
    to: to || undefined,
    sort,
    page: params.page,
    pageSize: params.pageSize,
  })

  if (hasErrorCode(events.error, 'forbidden_field')) return null

  const typeNames = new Map((eventTypes.data ?? []).map((item) => [item.code, item.name]))

  const columns: Column<GuestEvent>[] = [
    {
      id: 'occurredAt',
      header: 'Дата',
      cell: (e) => formatDateTime(e.occurredAt),
      sortable: true,
    },
    ...(purchasesOnly
      ? []
      : [
          {
            id: 'type',
            header: 'Событие',
            cell: (e: GuestEvent) =>
              typeNames.get(e.eventTypeCode ?? '') ?? e.eventTypeCode ?? 'Событие',
          },
        ]),
    { id: 'tenant', header: 'Арендатор', cell: (e) => e.tenantName ?? e.tenantId },
    ...(canSeeAmount
      ? [
          {
            id: 'amountMinor',
            header: 'Сумма',
            cell: (e: GuestEvent) => (e.amountMinor == null ? '—' : formatMoney(e.amountMinor)),
            className: 'text-right',
            sortable: true,
          },
        ]
      : []),
  ]

  const typeItems = [
    { value: ALL_OPTION, label: 'Все события' },
    ...(eventTypes.data ?? []).map((item) => ({ value: item.code, label: item.name })),
  ]
  const tenantItems = [
    { value: ALL_OPTION, label: 'Все арендаторы' },
    ...tenants.map((item) => ({ value: item.tenantId, label: item.tenantName })),
  ]
  const fromSelect = (value: string) => (value === ALL_OPTION ? '' : value)

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-start gap-3">
        {!purchasesOnly && (
          <OptionSelect
            ariaLabel="Тип события"
            className="mt-5 w-52"
            items={typeItems}
            value={type || ALL_OPTION}
            onValueChange={(value) => params.update({ type: fromSelect(value) })}
          />
        )}
        <OptionSelect
          ariaLabel="Арендатор"
          className="mt-5 w-56"
          items={tenantItems}
          value={tenantId || ALL_OPTION}
          onValueChange={(value) => params.update({ tenant_id: fromSelect(value) })}
        />
        <PeriodFilter
          value={{ from, to }}
          onChange={(period) => params.update({ from: period.from, to: period.to })}
          error={invalidPeriod ? PERIOD_REVERSED_MESSAGE : undefined}
        >
          <Button
            type="button"
            variant="outline"
            className="mt-5"
            disabled={!isFiltered}
            onClick={() => params.update({ type: null, tenant_id: null, from: null, to: null })}
          >
            Сбросить
          </Button>
        </PeriodFilter>
      </div>
      <DataTable
        columns={columns}
        rows={events.data?.items}
        getRowId={(e) => e.id}
        isLoading={events.isLoading}
        error={events.error}
        emptyTitle={
          isFiltered
            ? 'По выбранным условиям ничего не найдено'
            : purchasesOnly
              ? 'Покупок пока нет'
              : 'Событий пока нет'
        }
        meta={events.data?.meta}
        onPageChange={params.setPage}
        onPageSizeChange={params.setPageSize}
        sort={shownSort}
        onSortChange={(next) =>
          params.setSort(next ?? { field: shownSort.field, direction: 'asc' })
        }
      />
    </div>
  )
}
