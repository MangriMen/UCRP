import { useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router'
import { useDataSources, type DataSource } from '@/entities/data-source'
import { useDataSourceTypes } from '@/entities/data-source-type'
import {
  ImportStatusBadge,
  importStatusLabels,
  isImportActive,
  useImports,
} from '@/entities/import-batch'
import { useAuth } from '@/entities/user'
import { DataSourceFormDialog, DeleteDataSourceDialog } from '@/features/manage-data-sources'
import { SyncSourceButton } from '@/features/sync-source'
import { formatDateTime } from '@/shared/lib/format'
import { ALL_OPTION } from '@/shared/lib/search-params'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { Input } from '@/shared/ui/input'
import { OptionSelect } from '@/shared/ui/option-select'
import { PageHeader } from '@/shared/ui/page-header'
import { latestBatchBySource } from '../lib/latest-batch'
import { sourceKindLabel } from '../lib/source-kind'

const NO_IMPORTS = 'none'
const IMPORTS_PAGE_SIZE = 200

interface FormState {
  source: DataSource | null
}

export function DataSourcesPage() {
  const { can } = useAuth()
  const canEdit = can('data_source.write')
  const canImport = can('import.create')
  const [searchParams, setSearchParams] = useSearchParams()
  const typeId = searchParams.get('type') ?? ALL_OPTION
  const status = searchParams.get('status') ?? ALL_OPTION
  const query = searchParams.get('q') ?? ''

  const sources = useDataSources()
  const types = useDataSourceTypes()
  const imports = useImports({ pageSize: IMPORTS_PAGE_SIZE })
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<DataSource | null>(null)

  const updateParams = (changes: Record<string, string>) =>
    setSearchParams(
      (current) => {
        const next = new URLSearchParams(current)
        for (const [key, value] of Object.entries(changes)) {
          if (value === '' || value === ALL_OPTION) next.delete(key)
          else next.set(key, value)
        }
        return next
      },
      { replace: true },
    )

  const typeById = new Map((types.data ?? []).map((t) => [t.id, t]))
  const latest = useMemo(() => latestBatchBySource(imports.data?.items ?? []), [imports.data])
  const search = query.trim().toLowerCase()

  const rows = useMemo(
    () =>
      (sources.data ?? []).filter((source) => {
        const batchStatus = latest.get(source.id)?.status ?? NO_IMPORTS
        return (
          (typeId === ALL_OPTION || source.typeId === typeId) &&
          (status === ALL_OPTION || batchStatus === status) &&
          (search === '' || (source.tenantName ?? '').toLowerCase().includes(search))
        )
      }),
    [sources.data, latest, typeId, status, search],
  )

  const typeItems = [
    { value: ALL_OPTION, label: 'Все типы' },
    ...(types.data ?? []).map((t) => ({ value: t.id, label: t.name })),
  ]
  const statusItems = [
    { value: ALL_OPTION, label: 'Все статусы' },
    ...Object.entries(importStatusLabels).map(([value, label]) => ({ value, label })),
    { value: NO_IMPORTS, label: 'Нет импортов' },
  ]

  const columns: Column<DataSource>[] = [
    { id: 'tenant', header: 'Арендатор', cell: (s) => s.tenantName ?? 'Общий источник' },
    { id: 'name', header: 'Источник', cell: (s) => s.name },
    {
      id: 'type',
      header: 'Тип',
      cell: (s) => sourceKindLabel(s.typeCode, typeById.get(s.typeId)?.name),
    },
    {
      id: 'lastImportAt',
      header: 'Последний импорт',
      cell: (s) => (s.lastImportAt ? formatDateTime(s.lastImportAt) : '—'),
    },
    {
      id: 'status',
      header: 'Статус импорта',
      cell: (s) => {
        const batch = latest.get(s.id)
        return batch ? <ImportStatusBadge status={batch.status} /> : '—'
      },
    },
    { id: 'active', header: 'Активность', cell: (s) => <ActiveBadge isActive={s.isActive} /> },
    {
      id: 'actions',
      header: '',
      className: 'text-right',
      cell: (s) => {
        const type = typeById.get(s.typeId)
        const batch = latest.get(s.id)
        return (
          <div className="flex items-start justify-end gap-2">
            {canEdit && type?.supportsPull && s.isActive && (
              <SyncSourceButton
                sourceId={s.id}
                isRunning={batch !== undefined && isImportActive(batch.status)}
              />
            )}
            {canImport && type?.supportsFile && (
              <Button variant="outline" size="sm" render={<Link to={`/import?source=${s.id}`} />}>
                Загрузить файл
              </Button>
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
        description="Источники данных арендаторов и статус их последней синхронизации"
        actions={
          canEdit && <Button onClick={() => setForm({ source: null })}>Добавить источник</Button>
        }
      />
      <div className="flex flex-wrap gap-2">
        <Input
          className="w-56"
          aria-label="Поиск по арендатору"
          placeholder="Поиск по арендатору"
          value={query}
          onChange={(event) => updateParams({ q: event.target.value })}
        />
        <OptionSelect
          className="w-56"
          ariaLabel="Тип источника"
          items={typeItems}
          value={typeId}
          onValueChange={(value) => updateParams({ type: value })}
        />
        <OptionSelect
          className="w-56"
          ariaLabel="Статус импорта"
          items={statusItems}
          value={status}
          onValueChange={(value) => updateParams({ status: value })}
        />
      </div>
      <DataTable
        columns={columns}
        rows={rows}
        getRowId={(s) => s.id}
        isLoading={sources.isLoading}
        error={sources.error}
        emptyTitle="Источников не найдено"
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
