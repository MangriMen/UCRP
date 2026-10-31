import { useState } from 'react'
import { contactChannelLabels, useGuestContacts, type GuestContact } from '@/entities/guest'
import { formatDateTime } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { isFieldForbidden } from '../lib/is-field-forbidden'

const columns: Column<GuestContact>[] = [
  { id: 'contactedAt', header: 'Дата', cell: (c) => formatDateTime(c.contactedAt) },
  { id: 'channel', header: 'Канал', cell: (c) => contactChannelLabels[c.channel] },
]

export function ContactHistory({ guestId }: { guestId: string }) {
  const [page, setPage] = useState(1)
  const contacts = useGuestContacts(guestId, { page })

  if (isFieldForbidden(contacts.error)) return null

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
      />
    </section>
  )
}
