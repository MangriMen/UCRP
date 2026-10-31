import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'
import { GuestStatusBadge, MIN_SEARCH_LENGTH, useGuestSearch, type Guest } from '@/entities/guest'
import { useDebouncedValue } from '@/shared/lib/hooks/use-debounced-value'
import { formatDateTime } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { EmptyState } from '@/shared/ui/empty-state'
import { Input } from '@/shared/ui/input'
import { PageHeader } from '@/shared/ui/page-header'

const SEARCH_DEBOUNCE_MS = 300

const columns: Column<Guest>[] = [
  { id: 'name', header: 'Гость', cell: (g) => g.displayName ?? 'Без имени' },
  { id: 'status', header: 'Статус', cell: (g) => <GuestStatusBadge status={g.status} /> },
  { id: 'tenants', header: 'Арендаторов', cell: (g) => g.tenantCount, className: 'text-right' },
  { id: 'events', header: 'Событий', cell: (g) => g.eventCount ?? 0, className: 'text-right' },
  {
    id: 'lastSeen',
    header: 'Последняя активность',
    cell: (g) => (g.lastSeenAt ? formatDateTime(g.lastSeenAt) : '—'),
  },
]

const parsePage = (value: string | null) => {
  const page = Number(value)
  return Number.isInteger(page) && page > 0 ? page : 1
}

export function GuestsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const urlQuery = searchParams.get('q') ?? ''
  const page = parsePage(searchParams.get('page'))

  const [input, setInput] = useState(urlQuery)
  const query = useDebouncedValue(input.trim(), SEARCH_DEBOUNCE_MS)

  useEffect(() => {
    if (query === urlQuery) return
    setSearchParams((current) => {
      const next = new URLSearchParams(current)
      if (query) next.set('q', query)
      else next.delete('q')
      next.delete('page')
      return next
    })
  }, [query, urlQuery, setSearchParams])

  const guests = useGuestSearch({ query: urlQuery, page })
  const canSearch = urlQuery.length >= MIN_SEARCH_LENGTH

  const changePage = (next: number) =>
    setSearchParams((current) => {
      const params = new URLSearchParams(current)
      params.set('page', String(next))
      return params
    })

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Гости"
        description="Единая база профилей: поиск по телефону, email, карте или имени"
      />
      <Input
        value={input}
        onChange={(event) => setInput(event.target.value)}
        placeholder="Телефон, email, номер карты или имя"
        aria-label="Поиск гостей"
        className="max-w-md"
      />
      {canSearch ? (
        <DataTable
          columns={columns}
          rows={guests.data?.items}
          getRowId={(g) => g.id}
          isLoading={guests.isLoading}
          error={guests.error}
          emptyTitle="Гости не найдены"
          meta={guests.data?.meta}
          onPageChange={changePage}
        />
      ) : (
        <EmptyState
          title="Введите запрос для поиска"
          description={`Минимум ${MIN_SEARCH_LENGTH} символа`}
        />
      )}
    </div>
  )
}
