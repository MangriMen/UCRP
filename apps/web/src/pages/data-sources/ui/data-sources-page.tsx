import { useState } from 'react'
import { useDataSources, type DataSource } from '@/entities/data-source'
import { useDataSourceTypes } from '@/entities/data-source-type'
import { ImportStatusBadge, isImportActive } from '@/entities/import-batch'
import { useAuth } from '@/entities/user'
import { DataSourceFormDialog, DeleteDataSourceDialog } from '@/features/manage-data-sources'
import { SyncSourceButton } from '@/features/sync-source'
import { formatDateTime } from '@/shared/lib/format'
import { formatSort, useListParams } from '@/shared/lib/list-params'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button, ButtonLink } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'
import { sourceKindLabel } from '../lib/source-kind'
import { DATA_SOURCE_FILTER_KEYS } from '../model/filter-keys'
import { SourceFilters } from './source-filters'

const SORTABLE_FIELDS = ['tenantName', 'name', 'typeName', 'lastImportAt']

interface FormState {
  source: DataSource | null
}

export function DataSourcesPage() {
  const { can } = useAuth()
  const canEdit = can('data_source.write')
  const canImport = can('import.create')
  const params = useListParams()
  const { get, page, pageSize, setPage, setPageSize, setSort, reset } = params
  const isActive = get('is_active')
  const isFiltered = DATA_SOURCE_FILTER_KEYS.some((key) => get(key) !== '')
  const sort = params.sort && SORTABLE_FIELDS.includes(params.sort.field) ? params.sort : undefined

  const sources = useDataSources({
    q: get('q') || undefined,
    typeId: get('type_id') || undefined,
    importStatus: get('import_status') || undefined,
    isActive: isActive ? isActive === 'true' : undefined,
    sort: formatSort(sort),
    page,
    pageSize,
  })
  const types = useDataSourceTypes()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<DataSource | null>(null)

  const typeById = new Map((types.data ?? []).map((t) => [t.id, t]))

  const columns: Column<DataSource>[] = [
    {
      id: 'tenantName',
      header: 'Арендатор',
      cell: (s) => s.tenantName ?? 'Общий источник',
      sortable: true,
    },
    { id: 'name', header: 'Источник', cell: (s) => s.name, sortable: true },
    {
      id: 'typeName',
      header: 'Тип',
      sortable: true,
      cell: (s) => sourceKindLabel(s.typeCode, typeById.get(s.typeId)?.name),
    },
    {
      id: 'lastImportAt',
      header: 'Последняя загрузка',
      cell: (s) => (s.lastImportAt ? formatDateTime(s.lastImportAt) : '—'),
      sortable: true,
    },
    {
      id: 'status',
      header: 'Статус загрузки',
      cell: (s) => {
        return s.lastImport ? <ImportStatusBadge status={s.lastImport.status} /> : '—'
      },
    },
    { id: 'active', header: 'Активность', cell: (s) => <ActiveBadge isActive={s.isActive} /> },
    {
      id: 'actions',
      header: '',
      className: 'text-right',
      cell: (s) => {
        const type = typeById.get(s.typeId)
        return (
          <div className="flex items-start justify-end gap-2">
            {canEdit && type?.supportsPull && s.isActive && (
              <SyncSourceButton
                sourceId={s.id}
                isRunning={s.lastImport != null && isImportActive(s.lastImport.status)}
              />
            )}
            {canImport && type?.supportsFile && (
              <ButtonLink to={`/import?source=${s.id}`} variant="outline" size="sm">
                Загрузить файл
              </ButtonLink>
            )}
            {canEdit && (
              <>
                <Button variant="outline" size="sm" onClick={() => setForm({ source: s })}>
                  Изменить
                </Button>
                <Button variant="outline" size="sm" onClick={() => setDeleting(s)}>
                  Удалить
                </Button>
              </>
            )}
          </div>
        )
      },
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Источники данных"
        description="Источники данных арендаторов и статус их последней загрузки"
        actions={
          canEdit && <Button onClick={() => setForm({ source: null })}>Добавить источник</Button>
        }
      />
      <SourceFilters params={params} isFiltered={isFiltered} onReset={reset} />
      <DataTable
        columns={columns}
        rows={sources.data?.items}
        getRowId={(s) => s.id}
        isLoading={sources.isLoading}
        error={sources.error}
        emptyTitle={isFiltered ? 'Источники по фильтрам не найдены' : 'Источников пока нет'}
        emptyAction={
          isFiltered ? (
            <Button type="button" variant="outline" onClick={reset}>
              Сбросить фильтры
            </Button>
          ) : undefined
        }
        meta={sources.data?.meta}
        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        sort={sort}
        onSortChange={setSort}
      />
      {form && (
        <DataSourceFormDialog
          open
          source={form.source}
          onOpenChange={(open) => !open && setForm(null)}
        />
      )}
      <DeleteDataSourceDialog source={deleting} onClose={() => setDeleting(null)} />
    </div>
  )
}
