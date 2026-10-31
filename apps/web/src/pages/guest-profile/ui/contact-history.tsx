import { contactChannelLabels, useGuestContacts, type GuestContact } from '@/entities/guest'
import { hasErrorCode } from '@/shared/api/errors'
import { formatDateTime } from '@/shared/lib/format'
import { useListParams } from '@/shared/lib/list-params'
import { DataTable, type Column } from '@/shared/ui/data-table'

const columns: Column<GuestContact>[] = [
  { id: 'contactedAt', header: 'Дата', cell: (c) => formatDateTime(c.contactedAt) },
  { id: 'channel', header: 'Канал', cell: (c) => contactChannelLabels[c.channel] },
]

export function ContactHistory({ guestId }: { guestId: string }) {
  const { page, pageSize, setPage, setPageSize } = useListParams()
  const contacts = useGuestContacts(guestId, { page, pageSize })

  if (hasErrorCode(contacts.error, 'forbidden_field')) return null

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">История контактов</h2>
      <DataTable
        columns={columns}
        rows={contacts.data?.items}
        getRowId={(c) => c.id}
        isLoading={contacts.isLoading}
        error={contacts.error}
        emptyTitle="Контактов пока не было"
        meta={contacts.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
      />
    </section>
  )
}
