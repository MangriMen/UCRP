import { useEffect, useRef, useState } from 'react'
import { useMalls } from '@/entities/mall'
import { useTenantCategories } from '@/entities/tenant-category'
import type { ListParams } from '@/shared/lib/list-params'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { OptionSelect, type Option } from '@/shared/ui/option-select'
import { PeriodFilter } from '@/shared/ui/period-filter'

const SEARCH_DEBOUNCE_MS = 300

const activityOptions: Option[] = [
  { value: ALL_OPTION, label: 'Любая активность' },
  { value: 'true', label: 'Активные' },
  { value: 'false', label: 'Неактивные' },
]

interface TenantFiltersProps {
  params: ListParams
  isFiltered: boolean
  onReset: () => void
}

export function TenantFilters({ params, isFiltered, onReset }: TenantFiltersProps) {
  const { get, update } = params
  const [input, setInput] = useState(get('q'))
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const malls = useMalls()
  const categories = useTenantCategories()

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

  const mallItems: Option[] = [
    { value: ALL_OPTION, label: 'Все ТЦ' },
    ...(malls.data ?? []).map((mall) => ({ value: mall.id, label: mall.name })),
  ]
  const categoryItems: Option[] = [
    { value: ALL_OPTION, label: 'Все категории' },
    ...(categories.data ?? []).map((category) => ({ value: category.id, label: category.name })),
  ]
  const selectValue = (key: string) => get(key) || ALL_OPTION
  const changeSelect = (key: string) => (value: string) =>
    update({ [key]: value === ALL_OPTION ? null : value })

  const periodFrom = get('period_from')
  const periodTo = get('period_to')
  const invalidPeriod = periodFrom !== '' && periodTo !== '' && periodFrom > periodTo

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-3">
        <Input
          value={input}
          onChange={(event) => changeInput(event.target.value)}
          placeholder="Название или код"
          aria-label="Поиск арендаторов"
          className="max-w-xs"
        />
        <OptionSelect
          ariaLabel="ТЦ"
          className="w-48"
          items={mallItems}
          value={selectValue('mall_id')}
          onValueChange={changeSelect('mall_id')}
        />
        <OptionSelect
          ariaLabel="Категория"
          className="w-56"
          items={categoryItems}
          value={selectValue('category_id')}
          onValueChange={changeSelect('category_id')}
        />
        <Input
          type="number"
          inputMode="numeric"
          value={get('floor')}
          onChange={(event) => update({ floor: event.target.value })}
          placeholder="Этаж"
          aria-label="Этаж"
          className="w-24"
        />
        <OptionSelect
          ariaLabel="Активность"
          className="w-48"
          items={activityOptions}
          value={selectValue('is_active')}
          onValueChange={changeSelect('is_active')}
        />
      </div>
      <PeriodFilter
        value={{ from: periodFrom, to: periodTo }}
        onChange={({ from, to }) => update({ period_from: from, period_to: to })}
        error={invalidPeriod ? 'Начало периода позже конца' : undefined}
      >
        <div className="mt-5">
          <Button type="button" variant="outline" disabled={!isFiltered} onClick={reset}>
            Сбросить
          </Button>
        </div>
      </PeriodFilter>
    </div>
  )
}
