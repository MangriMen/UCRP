import { useMemo } from 'react'
import { Link } from 'react-router'
import { useDataSources, type DataSource } from '@/entities/data-source'
import { useDataSourceTypes } from '@/entities/data-source-type'
import { ImportStatusBadge, latestBatchBySource, useImports } from '@/entities/import-batch'
import { formatDateTime } from '@/shared/lib/format'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'

const IMPORTS_PAGE_SIZE = 200

export function SourcesTab({ tenantId }: { tenantId: string }) {
  const sources = useDataSources({ tenantId })
  const types = useDataSourceTypes()
  const imports = useImports({ pageSize: IMPORTS_PAGE_SIZE })
  const typeById = new Map((types.data ?? []).map((type) => [type.id, type]))
  const latest = useMemo(() => latestBatchBySource(imports.data?.items ?? []), [imports.data])

  const columns: Column<DataSource>[] = [
    { id: 'name', header: 'Источник', cell: (s) => s.name },
    { id: 'type', header: 'Тип', cell: (s) => typeById.get(s.typeId)?.name ?? s.typeCode },
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
  ]

  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-end">
        <Button variant="outline" render={<Link to="/data-sources" />}>
          Все источники данных
        </Button>
      </div>
      <DataTable
        columns={columns}
        rows={sources.data}
        getRowId={(s) => s.id}
        isLoading={sources.isLoading}
        error={sources.error}
        emptyTitle="У арендатора нет источников данных"
      />
    </div>
  )
}
