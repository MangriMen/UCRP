import { useState } from 'react'
import { useDataSourceTypes, type DataSourceType } from '@/entities/data-source-type'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { DataSourceTypeFormDialog } from './data-source-type-form-dialog'

interface DataSourceTypesManagerProps {
  canEdit: boolean
}

interface FormState {
  type: DataSourceType | null
}

const yesNo = (value: boolean) => (value ? 'Да' : 'Нет')

export function DataSourceTypesManager({ canEdit }: DataSourceTypesManagerProps) {
  const types = useDataSourceTypes()
  const [form, setForm] = useState<FormState | null>(null)

  const columns: Column<DataSourceType>[] = [
    { id: 'code', header: 'Код', cell: (t) => t.code },
    { id: 'name', header: 'Название', cell: (t) => t.name },
    { id: 'connector', header: 'Коннектор', cell: (t) => t.connectorKey },
    { id: 'file', header: 'Файлы', cell: (t) => yesNo(t.supportsFile) },
    { id: 'pull', header: 'API', cell: (t) => yesNo(t.supportsPull) },
    { id: 'status', header: 'Статус', cell: (t) => <ActiveBadge isActive={t.isActive ?? true} /> },
    ...(canEdit
      ? [
          {
            id: 'actions',
            header: '',
            className: 'text-right',
            cell: (t: DataSourceType) => (
              <div className="flex justify-end">
                <Button variant="outline" size="sm" onClick={() => setForm({ type: t })}>
                  Изменить
                </Button>
              </div>
            ),
          },
        ]
      : []),
  ]

  return (
    <div className="flex flex-col gap-4">
      {canEdit && (
        <div className="flex justify-end">
          <Button onClick={() => setForm({ type: null })}>Добавить тип источника</Button>
        </div>
      )}
      <DataTable
        columns={columns}
        rows={types.data}
        getRowId={(t) => t.id}
        isLoading={types.isLoading}
        error={types.error}
        emptyTitle="Типов источников пока нет"
      />
      {form && (
        <DataSourceTypeFormDialog
          open
          type={form.type}
          onOpenChange={(open) => !open && setForm(null)}
        />
      )}
    </div>
  )
}
