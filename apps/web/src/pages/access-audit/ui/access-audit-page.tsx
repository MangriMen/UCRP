import { AccessAuditTable } from '@/entities/access'
import { useAuth } from '@/entities/user'
import { AuditActor, useSubjectDirectory } from '@/features/manage-access-grants'
import { useDebouncedSearchParam } from '@/shared/lib/hooks/use-debounced-search-param'
import { useListParams } from '@/shared/lib/list-params'
import { endOfDay, startOfDay } from '@/shared/lib/period'
import { ALL_OPTION, optionParam } from '@/shared/lib/search-params'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'
import { PeriodFilter } from '@/shared/ui/period-filter'

const FILTER_KEYS = ['actor', 'grant_id', 'from', 'to'] as const

export function AccessAuditPage() {
  const { can } = useAuth()
  const directory = useSubjectDirectory()
  const params = useListParams()
  const { get, update, page, pageSize, setPage, setPageSize, reset } = params
  const grant = useDebouncedSearchParam(params, 'grant_id')

  const from = get('from')
  const to = get('to')
  const invalidPeriod = from !== '' && to !== '' && from > to
  const isFiltered = params.isFiltered(FILTER_KEYS)

  const resetFilters = () => {
    grant.clear()
    reset()
  }

  const actorItems = [
    { value: ALL_OPTION, label: 'Все пользователи' },
    ...directory.users.map(({ id, label }) => ({ value: id, label })),
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Аудит доступа"
        description="Выдача, изменение и отзыв грантов, проверки доступа и отказы"
        breadcrumbs={[
          { label: 'Доступы', to: can('access.grant.read') ? '/access/grants' : undefined },
          { label: 'Аудит доступа' },
        ]}
      />
      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-3">
          <OptionSelect
            ariaLabel="Кто"
            className="w-64"
            items={actorItems}
            value={get('actor') || ALL_OPTION}
            onValueChange={(value) => update({ actor: optionParam(value) })}
          />
          <Input
            aria-label="ID гранта"
            placeholder="ID гранта"
            className="w-64"
            value={grant.value}
            onChange={(event) => grant.onChange(event.target.value)}
          />
        </div>
        <PeriodFilter
          value={{ from, to }}
          onChange={(next) => update({ from: next.from, to: next.to })}
          error={invalidPeriod ? 'Начало периода позже конца' : undefined}
        >
          <div className="mt-5">
            <Button type="button" variant="outline" disabled={!isFiltered} onClick={resetFilters}>
              Сбросить
            </Button>
          </div>
        </PeriodFilter>
      </div>
      <AccessAuditTable
        filters={{
          actor: get('actor') || undefined,
          grantId: get('grant_id') || undefined,
          from: from && !invalidPeriod ? startOfDay(from) : undefined,
          to: to && !invalidPeriod ? endOfDay(to) : undefined,
          page,
          pageSize,
        }}
        actorLabel={(entry) => <AuditActor entry={entry} directory={directory} />}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        emptyTitle={isFiltered ? 'Записи по фильтрам не найдены' : 'Записей пока нет'}
        onResetFilters={isFiltered ? resetFilters : undefined}
      />
    </div>
  )
}
