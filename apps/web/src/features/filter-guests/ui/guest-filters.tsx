import { useEffect, useRef, useState } from 'react'
import { guestStatusLabels } from '@/entities/guest'
import { useDataSources } from '@/entities/data-source'
import { useSegments } from '@/entities/segment'
import { useTenants } from '@/entities/tenant'
import type { ListParams } from '@/shared/lib/list-params'
import { daysAgo } from '@/shared/lib/period'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { OptionSelect, type Option } from '@/shared/ui/option-select'
import { PeriodFilter } from '@/shared/ui/period-filter'

const SEARCH_DEBOUNCE_MS = 300
const OPTIONS_LIMIT = 200
const NEW_GUESTS_DAYS = 7
const INACTIVE_GUESTS_DAYS = 90

const purchaseOptions: Option[] = [
  { value: ALL_OPTION, label: 'Любые покупки' },
  { value: 'true', label: 'Есть покупки' },
  { value: 'false', label: 'Без покупок' },
]

const statusOptions: Option[] = [
  { value: ALL_OPTION, label: 'Активные' },
  ...Object.entries(guestStatusLabels)
    .filter(([status]) => status !== 'active')
    .map(([value, label]) => ({ value, label })),
]

interface GuestFiltersPanelProps {
  params: ListParams
  isFiltered: boolean
  onReset: () => void
}

const withAll = (label: string, items: Option[]): Option[] => [
  { value: ALL_OPTION, label },
  ...items,
]

export function GuestFiltersPanel({ params, isFiltered, onReset }: GuestFiltersPanelProps) {
  const { get, update } = params
  const [input, setInput] = useState(get('q'))
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const sources = useDataSources()
  const tenants = useTenants({ pageSize: OPTIONS_LIMIT })
  const segments = useSegments({ pageSize: OPTIONS_LIMIT })

  useEffect(() => () => clearTimeout(timer.current), [])

  const changeInput = (value: string) => {
    setInput(value)
    clearTimeout(timer.current)
    timer.current = setTimeout(() => update({ q: value.trim() }), SEARCH_DEBOUNCE_MS)
  }

  const reset = () => {
    clearTimeout(timer.current)
    setInput('')
    onReset()
  }

  const sourceItems = withAll(
    'Все источники',
    (sources.data ?? []).map((source) => ({ value: source.id, label: source.name })),
  )
  const tenantItems = withAll(
    'Все арендаторы',
    (tenants.data?.items ?? []).map((tenant) => ({ value: tenant.id, label: tenant.name })),
  )
  const segmentItems = withAll(
    'Все сегменты',
    (segments.data?.items ?? []).map((segment) => ({ value: segment.id, label: segment.name })),
  )
  const selectValue = (key: string) => get(key) || ALL_OPTION
  const changeSelect = (key: string) => (value: string) =>
    update({ [key]: value === ALL_OPTION ? null : value })

  const periodFrom = get('period_from')
  const periodTo = get('period_to')
  const invalidPeriod = periodFrom !== '' && periodTo !== '' && periodFrom > periodTo

  const newSince = daysAgo(NEW_GUESTS_DAYS)
  const inactiveSince = daysAgo(INACTIVE_GUESTS_DAYS)
  const newActive = get('first_seen_from') === newSince && get('last_seen_to') === ''
  const inactiveActive = get('last_seen_to') === inactiveSince && get('first_seen_from') === ''

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-3">
        <Input
          value={input}
          onChange={(event) => changeInput(event.target.value)}
          placeholder="Телефон, email, номер карты или имя"
          aria-label="Поиск гостей"
          className="max-w-md"
        />
        <OptionSelect
          ariaLabel="Источник"
          className="w-56"
          items={sourceItems}
          value={selectValue('source_id')}
          onValueChange={changeSelect('source_id')}
        />
        <OptionSelect
          ariaLabel="Арендатор"
          className="w-56"
          items={tenantItems}
          value={selectValue('tenant_id')}
          onValueChange={changeSelect('tenant_id')}
        />
        <OptionSelect
          ariaLabel="Сегмент"
          className="w-56"
          items={segmentItems}
          value={selectValue('segment_id')}
          onValueChange={changeSelect('segment_id')}
        />
        <OptionSelect
          ariaLabel="Статус"
          className="w-40"
          items={statusOptions}
          value={selectValue('status')}
          onValueChange={changeSelect('status')}
        />
        <OptionSelect
          ariaLabel="Покупки"
          className="w-44"
          items={purchaseOptions}
          value={selectValue('has_purchases')}
          onValueChange={changeSelect('has_purchases')}
        />
      </div>
      <PeriodFilter
        value={{ from: periodFrom, to: periodTo }}
        onChange={({ from, to }) => update({ period_from: from, period_to: to })}
        error={invalidPeriod ? 'Начало периода позже конца' : undefined}
      >
        <div className="mt-5 flex flex-wrap gap-2">
          <Button
            type="button"
            variant={newActive ? 'secondary' : 'outline'}
            aria-pressed={newActive}
            onClick={() =>
              update({ first_seen_from: newActive ? null : newSince, last_seen_to: null })
            }
          >
            Новые за {NEW_GUESTS_DAYS} дней
          </Button>
          <Button
            type="button"
            variant={inactiveActive ? 'secondary' : 'outline'}
            aria-pressed={inactiveActive}
            onClick={() =>
              update({ last_seen_to: inactiveActive ? null : inactiveSince, first_seen_from: null })
            }
          >
            Неактивные {INACTIVE_GUESTS_DAYS}+ дней
          </Button>
          <Button type="button" variant="outline" disabled={!isFiltered} onClick={reset}>
            Сбросить
          </Button>
        </div>
      </PeriodFilter>
    </div>
  )
}
