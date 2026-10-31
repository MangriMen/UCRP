import { useState, type FormEvent } from 'react'
import { useContactLimits, type ContactLimit } from '@/entities/contact-limit'
import { exportChannelLabels, type ExportChannel } from '@/entities/export'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/shared/ui/select'
import {
  useCreateContactLimit,
  useDeleteContactLimit,
  useUpdateContactLimit,
} from '../api/use-contact-limit-mutations'
import { validateLimit } from '../model/validate-limit'

const ALL_CHANNELS = 'all'

const channelItems = [
  { value: ALL_CHANNELS, label: 'Все каналы' },
  ...Object.entries(exportChannelLabels).map(([value, label]) => ({ value, label })),
]

interface LimitSettingsProps {
  canEdit: boolean
}

export function LimitSettings({ canEdit }: LimitSettingsProps) {
  const limits = useContactLimits()
  const create = useCreateContactLimit()
  const update = useUpdateContactLimit()
  const remove = useDeleteContactLimit()
  const [editing, setEditing] = useState<ContactLimit | null>(null)
  const [channel, setChannel] = useState(ALL_CHANNELS)
  const [maxContacts, setMaxContacts] = useState('')
  const [periodDays, setPeriodDays] = useState('')
  const [deleting, setDeleting] = useState<ContactLimit | null>(null)

  const errors = validateLimit(maxContacts, periodDays)
  const isValid = !errors.maxContacts && !errors.periodDays
  const mutationError = create.error ?? update.error ?? remove.error

  const reset = () => {
    setEditing(null)
    setChannel(ALL_CHANNELS)
    setMaxContacts('')
    setPeriodDays('')
  }

  const startEdit = (limit: ContactLimit) => {
    setEditing(limit)
    setChannel(limit.channel ?? ALL_CHANNELS)
    setMaxContacts(String(limit.maxContacts))
    setPeriodDays(String(limit.periodDays))
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!isValid) return
    const body = {
      channel: channel === ALL_CHANNELS ? null : (channel as ExportChannel),
      maxContacts: Number(maxContacts),
      periodDays: Number(periodDays),
    }
    if (editing) update.mutate({ id: editing.id, body }, { onSuccess: reset })
    else create.mutate(body, { onSuccess: reset })
  }

  const columns: Column<ContactLimit>[] = [
    {
      id: 'channel',
      header: 'Канал',
      cell: (l) => (l.channel ? exportChannelLabels[l.channel] : 'Все каналы'),
    },
    { id: 'max', header: 'Макс. контактов', cell: (l) => l.maxContacts, className: 'text-right' },
    { id: 'period', header: 'За период, дней', cell: (l) => l.periodDays, className: 'text-right' },
    ...(canEdit
      ? [
          {
            id: 'actions',
            header: '',
            className: 'text-right',
            cell: (l: ContactLimit) => (
              <div className="flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={() => startEdit(l)}>
                  Изменить
                </Button>
                <Button variant="outline" size="sm" onClick={() => setDeleting(l)}>
                  Удалить
                </Button>
              </div>
            ),
          },
        ]
      : []),
  ]

  return (
    <div className="flex flex-col gap-6">
      {mutationError && <ErrorBanner error={mutationError} />}
      <DataTable
        columns={columns}
        rows={limits.data}
        getRowId={(l) => l.id}
        isLoading={limits.isLoading}
        error={limits.error}
        emptyTitle="Лимиты не заданы"
      />
      {canEdit && (
        <Card>
          <CardHeader>
            <CardTitle>{editing ? 'Изменить лимит' : 'Новый лимит'}</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={submit} className="flex flex-wrap items-start gap-4">
              <FormField label="Канал">
                {(field) => (
                  <Select
                    items={channelItems}
                    value={channel}
                    onValueChange={(value) => value && setChannel(value)}
                  >
                    <SelectTrigger {...field} className="w-44">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {channelItems.map((item) => (
                        <SelectItem key={item.value} value={item.value}>
                          {item.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField
                label="Макс. контактов"
                error={maxContacts ? errors.maxContacts : undefined}
              >
                {(field) => (
                  <Input
                    {...field}
                    type="number"
                    className="w-40"
                    value={maxContacts}
                    onChange={(event) => setMaxContacts(event.target.value)}
                  />
                )}
              </FormField>
              <FormField label="За период, дней" error={periodDays ? errors.periodDays : undefined}>
                {(field) => (
                  <Input
                    {...field}
                    type="number"
                    className="w-40"
                    value={periodDays}
                    onChange={(event) => setPeriodDays(event.target.value)}
                  />
                )}
              </FormField>
              <div className="flex gap-2 pt-6">
                <Button type="submit" disabled={!isValid || create.isPending || update.isPending}>
                  {editing ? 'Сохранить' : 'Добавить'}
                </Button>
                {editing && (
                  <Button type="button" variant="outline" onClick={reset}>
                    Отмена
                  </Button>
                )}
              </div>
            </form>
          </CardContent>
        </Card>
      )}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Удалить лимит?"
        description="Правило перестанет учитываться при новых выгрузках."
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
