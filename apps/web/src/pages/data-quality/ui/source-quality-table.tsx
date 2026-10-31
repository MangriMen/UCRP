import { warningReasonLabel, type DataQualitySource } from '@/entities/data-quality'
import { formatDateTime, formatPercent } from '@/shared/lib/format'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { DataTable, type Column } from '@/shared/ui/data-table'

interface SourceQualityTableProps {
  rows: DataQualitySource[] | undefined
  isLoading: boolean
  error: unknown
  onSelect: (id: string) => void
}

export function SourceQualityTable({ rows, isLoading, error, onSelect }: SourceQualityTableProps) {
  const columns: Column<DataQualitySource>[] = [
    { id: 'source', header: 'Источник', cell: (s) => s.dataSourceName },
    { id: 'tenant', header: 'Арендатор', cell: (s) => s.tenantName ?? 'Общий источник' },
    {
      id: 'lastImport',
      header: 'Последний успешный импорт',
      cell: (s) => (s.lastSuccessfulImportAt ? formatDateTime(s.lastSuccessfulImportAt) : '—'),
    },
    {
      id: 'rejected',
      header: 'Отклонено за 30 дней',
      cell: (s) => formatPercent(s.rejectedShare30d),
    },
    { id: 'pending', header: 'Спорные профили', cell: (s) => s.pendingMergeCandidates },
    {
      id: 'warning',
      header: 'Состояние',
      cell: (s) =>
        s.hasWarning ? (
          <div className="flex flex-col items-start gap-1">
            <Badge variant="destructive">Требует внимания</Badge>
            {s.warningReasons?.map((reason) => (
              <span key={reason} className="text-xs text-muted-foreground">
                {warningReasonLabel(reason)}
              </span>
            ))}
          </div>
        ) : (
          <Badge variant="secondary">В норме</Badge>
        ),
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
