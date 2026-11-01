import { useMalls } from '@/entities/mall'
import { useTenantCategories } from '@/entities/tenant-category'
import { useDebouncedSearchParam } from '@/shared/lib/hooks/use-debounced-search-param'
import type { ListParams } from '@/shared/lib/list-params'
import { useOptionParam } from '@/shared/lib/hooks/use-option-param'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { OptionSelect, type Option } from '@/shared/ui/option-select'
import { PeriodFilter } from '@/shared/ui/period-filter'

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
  const search = useDebouncedSearchParam(params, 'q')
  const malls = useMalls()
  const categories = useTenantCategories()
  const mallParam = useOptionParam(
    params,
    'mall_id',
    malls.isSuccess ? malls.data.map((mall) => mall.id) : undefined,
  )
  const categoryParam = useOptionParam(
    params,
    'category_id',
    categories.isSuccess ? categories.data.map((category) => category.id) : undefined,
  )
  const activityParam = useOptionParam(
    params,
    'is_active',
    activityOptions.map((option) => option.value),
  )

  const reset = () => {
    search.clear()
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

  const periodFrom = get('period_from')
  const periodTo = get('period_to')
  const invalidPeriod = periodFrom !== '' && periodTo !== '' && periodFrom > periodTo

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-3">
        <Input
          value={search.value}
          onChange={(event) => search.onChange(event.target.value)}
          placeholder="Название или код"
          aria-label="Поиск арендаторов"
          className="max-w-xs"
        />
        <OptionSelect
          ariaLabel="ТЦ"
          className="w-48"
          items={mallItems}
          value={mallParam.value}
          onValueChange={mallParam.onChange}
        />
        <OptionSelect
          ariaLabel="Категория"
          className="w-56"
          items={categoryItems}
          value={categoryParam.value}
          onValueChange={categoryParam.onChange}
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
          value={activityParam.value}
          onValueChange={activityParam.onChange}
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
