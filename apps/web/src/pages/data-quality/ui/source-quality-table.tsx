import type { DataQualityListItem } from '@/entities/data-quality'
import { formatDateTime, formatPercent } from '@/shared/lib/format'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'

const suppressed = '—'

interface SourceQualityTableProps {
  rows: DataQualityListItem[] | undefined
  isLoading: boolean
  error: unknown
  onSelect: (id: string) => void
}

export function SourceQualityTable({ rows, isLoading, error, onSelect }: SourceQualityTableProps) {
  const columns: Column<DataQualityListItem>[] = [
    { id: 'source', header: 'Источник', cell: (s) => s.name },
    {
      id: 'lastImport',
      header: 'Последняя загрузка',
      cell: (s) => (s.lastImportAt ? formatDateTime(s.lastImportAt) : '—'),
    },
    {
      id: 'rejected',
      header: 'Отклонено за 30 дней',
      cell: (s) => (s.rejectedShare30d === null ? suppressed : formatPercent(s.rejectedShare30d)),
    },
    {
      id: 'pending',
      header: 'Спорные объединения',
      cell: (s) => s.pendingMergeCandidates ?? suppressed,
    },
    {
      id: 'warning',
      header: 'Состояние',
      cell: (s) => {
        if (s.suppressed) return <Badge variant="outline">Мало данных</Badge>
        return s.hasWarning ? (
          <Badge variant="destructive">Требует внимания</Badge>
        ) : (
          <Badge variant="secondary">В норме</Badge>
        )
      },
    },
    {
      id: 'actions',
      header: '',
      className: 'text-right',
      cell: (s) => (
        <Button variant="outline" size="sm" onClick={() => onSelect(s.dataSourceId)}>
          Подробнее
        </Button>
      ),
    },
  ]

  return (
    <DataTable
      columns={columns}
      rows={rows}
      getRowId={(s) => s.dataSourceId}
      isLoading={isLoading}
      error={error}
      emptyTitle="Источников нет"
    />
  )
}
