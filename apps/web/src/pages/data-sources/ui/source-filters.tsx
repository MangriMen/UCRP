import { useDataSourceTypes } from '@/entities/data-source-type'
import { importStatusLabels } from '@/entities/import-batch'
import { useDebouncedSearchParam } from '@/shared/lib/hooks/use-debounced-search-param'
import type { ListParams } from '@/shared/lib/list-params'
import { ALL_OPTION, optionParam } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { OptionSelect, type Option } from '@/shared/ui/option-select'

const statusItems: Option[] = [
  { value: ALL_OPTION, label: 'Все статусы' },
  ...Object.entries(importStatusLabels).map(([value, label]) => ({ value, label })),
  { value: 'none', label: 'Нет загрузок' },
]

const activityItems: Option[] = [
  { value: ALL_OPTION, label: 'Любая активность' },
  { value: 'true', label: 'Активные' },
  { value: 'false', label: 'Неактивные' },
]

interface SourceFiltersProps {
  params: ListParams
  isFiltered: boolean
  onReset: () => void
}

export function SourceFilters({ params, isFiltered, onReset }: SourceFiltersProps) {
  const { get, update } = params
  const search = useDebouncedSearchParam(params, 'q')
  const types = useDataSourceTypes()

  const reset = () => {
    search.clear()
    onReset()
  }

  const typeItems: Option[] = [
    { value: ALL_OPTION, label: 'Все типы' },
    ...(types.data ?? []).map((type) => ({ value: type.id, label: type.name })),
  ]
  const selectValue = (key: string) => get(key) || ALL_OPTION
  const changeSelect = (key: string) => (value: string) => update({ [key]: optionParam(value) })

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Input
        className="w-64"
        aria-label="Поиск по источнику или арендатору"
        placeholder="Источник или арендатор"
        value={search.value}
        onChange={(event) => search.onChange(event.target.value)}
      />
      <OptionSelect
        className="w-56"
        ariaLabel="Тип источника"
        items={typeItems}
        value={selectValue('type_id')}
        onValueChange={changeSelect('type_id')}
      />
      <OptionSelect
        className="w-56"
        ariaLabel="Статус загрузки"
        items={statusItems}
        value={selectValue('import_status')}
        onValueChange={changeSelect('import_status')}
      />
      <OptionSelect
        className="w-48"
        ariaLabel="Активность"
        items={activityItems}
        value={selectValue('is_active')}
        onValueChange={changeSelect('is_active')}
      />
      <Button type="button" variant="outline" disabled={!isFiltered} onClick={reset}>
        Сбросить
      </Button>
    </div>
  )
}
