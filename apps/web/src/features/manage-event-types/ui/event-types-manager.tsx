import { useState } from 'react'
import { useEventTypes, type EventType } from '@/entities/event-type'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteEventType } from '../api/use-event-type-mutations'
import { EventTypeFormDialog } from './event-type-form-dialog'

interface EventTypesManagerProps {
  canEdit: boolean
}

interface FormState {
  eventType: EventType | null
}

const yesNo = (value: boolean) => (value ? 'Да' : 'Нет')

export function EventTypesManager({ canEdit }: EventTypesManagerProps) {
  const eventTypes = useEventTypes()
  const remove = useDeleteEventType()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<EventType | null>(null)

  const columns: Column<EventType>[] = [
    { id: 'code', header: 'Код', cell: (t) => t.code },
    { id: 'name', header: 'Название', cell: (t) => t.name },
    { id: 'amount', header: 'Сумма', cell: (t) => yesNo(t.hasAmount) },
    { id: 'presence', header: 'Визит', cell: (t) => yesNo(t.countsAsPresence) },
    { id: 'status', header: 'Статус', cell: (t) => <ActiveBadge isActive={t.isActive} /> },
    ...(canEdit
      ? [
          {
            id: 'actions',
            header: '',
            className: 'text-right',
            cell: (t: EventType) => (
              <div className="flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={() => setForm({ eventType: t })}>
                  Изменить
                </Button>
                <Button variant="outline" size="sm" onClick={() => setDeleting(t)}>
                  Удалить
                </Button>
              </div>
            ),
          },
        ]
      : []),
  ]

  return (
    <div className="flex flex-col gap-4">
      {canEdit && (
        <div className="flex justify-end">
          <Button onClick={() => setForm({ eventType: null })}>Добавить тип события</Button>
        </div>
      )}
      {remove.error && <ErrorBanner error={remove.error} />}
      <DataTable
        columns={columns}
        rows={eventTypes.data}
        getRowId={(t) => t.id}
        isLoading={eventTypes.isLoading}
        error={eventTypes.error}
        emptyTitle="Типов событий пока нет"
      />
      {form && (
        <EventTypeFormDialog
          open
          eventType={form.eventType}
          onOpenChange={(open) => !open && setForm(null)}
        />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Удалить тип события?"
        description={`Тип «${deleting?.name ?? ''}» будет удалён. Если он используется в событиях, удаление будет отклонено.`}
        confirmLabel="Удалить"
        destructive
        isPending={remove.isPending}
        onConfirm={() => {
          if (deleting) remove.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
      />
    </div>
  )
}
