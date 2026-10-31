import { useState, type FormEvent } from 'react'
import { useSearchParams } from 'react-router'
import { useAuditLogs, type AuditLog } from '@/entities/audit-log'
import { roleLabels } from '@/entities/user'
import { formatDateTime } from '@/shared/lib/format'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { FormField } from '@/shared/ui/form-field'
import { Input } from '@/shared/ui/input'
import { PageHeader } from '@/shared/ui/page-header'

const FILTER_KEYS = ['entity_type', 'entity_id', 'actor', 'action', 'from', 'to'] as const

const startOfDay = (day: string) => `${day}T00:00:00Z`
const endOfDay = (day: string) => `${day}T23:59:59Z`

const parsePage = (value: string | null) => {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

const columns: Column<AuditLog>[] = [
  { id: 'createdAt', header: 'Время', cell: (l) => formatDateTime(l.createdAt) },
  {
    id: 'actor',
    header: 'Пользователь',
    cell: (l) => (
      <div>
        <div>{l.actorUserId ?? '—'}</div>
        {l.actorRole && (
          <div className="text-xs text-muted-foreground">{roleLabels[l.actorRole]}</div>
        )}
      </div>
    ),
  },
  { id: 'action', header: 'Действие', cell: (l) => l.action },
  { id: 'entity', header: 'Объект', cell: (l) => `${l.entityType} · ${l.entityId}` },
  {
    id: 'details',
    header: 'Детали',
    cell: (l) =>
      l.details ? <code className="text-xs break-all">{JSON.stringify(l.details)}</code> : '—',
  },
]

export function AuditLogPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [draft, setDraft] = useState(() =>
    Object.fromEntries(FILTER_KEYS.map((key) => [key, searchParams.get(key) ?? ''])),
  )

  const from = searchParams.get('from')
  const to = searchParams.get('to')
  const logs = useAuditLogs({
    entityType: searchParams.get('entity_type') ?? undefined,
    entityId: searchParams.get('entity_id') ?? undefined,
    actor: searchParams.get('actor') ?? undefined,
    action: searchParams.get('action') ?? undefined,
    from: from ? startOfDay(from) : undefined,
    to: to ? endOfDay(to) : undefined,
    page: parsePage(searchParams.get('page')),
  })

  const apply = (event: FormEvent) => {
    event.preventDefault()
    const next = new URLSearchParams()
    for (const key of FILTER_KEYS) {
      const value = draft[key].trim()
      if (value) next.set(key, value)
    }
    setSearchParams(next)
  }

  const reset = () => {
    setDraft(Object.fromEntries(FILTER_KEYS.map((key) => [key, ''])))
    setSearchParams(new URLSearchParams())
  }

  const changePage = (page: number) =>
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      next.set('page', String(page))
      return next
    })

  const textField = (key: (typeof FILTER_KEYS)[number], label: string, type = 'text') => (
    <FormField label={label}>
      {(field) => (
        <Input
          {...field}
          type={type}
          value={draft[key]}
          onChange={(event) => setDraft({ ...draft, [key]: event.target.value })}
          className="w-44"
        />
      )}
    </FormField>
  )

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Журнал аудита" description="Кто и что менял в системе" />
      <form onSubmit={apply} className="flex flex-wrap items-end gap-4">
        {textField('action', 'Действие')}
        {textField('entity_type', 'Тип объекта')}
        {textField('entity_id', 'ID объекта')}
        {textField('actor', 'ID пользователя')}
        {textField('from', 'С даты', 'date')}
        {textField('to', 'По дату', 'date')}
        <Button type="submit">Применить</Button>
        <Button type="button" variant="outline" onClick={reset}>
          Сбросить
        </Button>
      </form>
      <DataTable
        columns={columns}
        rows={logs.data?.items}
        getRowId={(l) => l.id}
        isLoading={logs.isLoading}
        error={logs.error}
        emptyTitle="Записей не найдено"
        meta={logs.data?.meta}
        onPageChange={changePage}
      />
    </div>
  )
}
