import { useEffect, useRef, useState } from 'react'
import { useDataSources } from '@/entities/data-source'
import { useTenants } from '@/entities/tenant'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { OptionSelect, type Option } from '@/shared/ui/option-select'
import { PeriodFilter } from '@/shared/ui/period-filter'
import type { GuestFilters } from '../model/use-guest-filters'

const SEARCH_DEBOUNCE_MS = 300
const TENANT_OPTIONS_LIMIT = 200

interface GuestFiltersPanelProps {
  filters: GuestFilters
  isFiltered: boolean
  onChange: (
    patch: Partial<Record<'q' | 'sourceId' | 'tenantId' | 'periodFrom' | 'periodTo', string>>,
  ) => void
  onReset: () => void
}

const withAll = (label: string, items: Option[]): Option[] => [
  { value: ALL_OPTION, label },
  ...items,
]

export function GuestFiltersPanel({
  filters,
  isFiltered,
  onChange,
  onReset,
}: GuestFiltersPanelProps) {
  const [input, setInput] = useState(filters.q)
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const sources = useDataSources()
  const tenants = useTenants({ pageSize: TENANT_OPTIONS_LIMIT })

  useEffect(() => () => clearTimeout(timer.current), [])

  const changeInput = (value: string) => {
    setInput(value)
    clearTimeout(timer.current)
    timer.current = setTimeout(() => onChange({ q: value.trim() }), SEARCH_DEBOUNCE_MS)
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
  const selectValue = (value: string) => value || ALL_OPTION
  const fromSelect = (value: string) => (value === ALL_OPTION ? '' : value)

  const invalidPeriod =
    filters.periodFrom !== '' && filters.periodTo !== '' && filters.periodFrom > filters.periodTo

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
          value={selectValue(filters.sourceId)}
          onValueChange={(value) => onChange({ sourceId: fromSelect(value) })}
        />
        <OptionSelect
          ariaLabel="Арендатор"
          className="w-56"
          items={tenantItems}
          value={selectValue(filters.tenantId)}
          onValueChange={(value) => onChange({ tenantId: fromSelect(value) })}
        />
      </div>
      <PeriodFilter
        value={{ from: filters.periodFrom, to: filters.periodTo }}
        onChange={({ from, to }) => onChange({ periodFrom: from, periodTo: to })}
        error={invalidPeriod ? 'Начало периода позже конца' : undefined}
      >
        <Button
          type="button"
          variant="outline"
          className="mt-5"
          disabled={!isFiltered}
          onClick={reset}
        >
          Сбросить
        </Button>
      </PeriodFilter>
    </div>
  )
}
