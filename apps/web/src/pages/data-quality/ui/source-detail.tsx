import { Link } from 'react-router'
import { useDataQualitySource, type DataQualityPoint } from '@/entities/data-quality'
import { ImportStatusBadge, type ImportBatch } from '@/entities/import-batch'
import { formatDate, formatDateTime, formatPercent } from '@/shared/lib/format'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Skeleton } from '@/shared/ui/skeleton'

const seriesColumns: Column<DataQualityPoint>[] = [
  { id: 'date', header: 'Дата', cell: (p) => formatDate(p.date) },
  { id: 'total', header: 'Строк', cell: (p) => p.totalRows ?? '—' },
  { id: 'rejected', header: 'Отклонено', cell: (p) => p.rejectedRows ?? '—' },
  {
    id: 'share',
    header: 'Доля',
    cell: (p) => (p.rejectedShare === undefined ? '—' : formatPercent(p.rejectedShare)),
  },
]

const batchColumns: Column<ImportBatch>[] = [
  { id: 'createdAt', header: 'Создан', cell: (b) => formatDateTime(b.createdAt) },
  { id: 'status', header: 'Статус', cell: (b) => <ImportStatusBadge status={b.status} /> },
  { id: 'rejected', header: 'Отклонено', cell: (b) => `${b.rejectedRows} из ${b.totalRows}` },
  {
    id: 'link',
    header: '',
    className: 'text-right',
    cell: (b) => (
      <Button variant="outline" size="sm" render={<Link to={`/import/${b.id}`} />}>
        Открыть импорт
      </Button>
    ),
  },
]

interface SourceDetailProps {
  sourceId: string
  onClose: () => void
}

export function SourceDetail({ sourceId, onClose }: SourceDetailProps) {
  const detail = useDataQualitySource(sourceId)

  if (detail.error) return <ErrorBanner error={detail.error} />
  if (!detail.data) return <Skeleton className="h-40 w-full" />

  const { dataSourceName, series, recentBatches } = detail.data
  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <CardTitle>{dataSourceName}</CardTitle>
        <Button variant="outline" size="sm" onClick={onClose}>
          Закрыть
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {series.length === 0 && recentBatches.length === 0 ? (
          <EmptyState title="Импортов пока не было" />
        ) : (
          <>
            <h3 className="font-medium">Динамика отклонённых строк</h3>
            <DataTable columns={seriesColumns} rows={series} getRowId={(p) => p.date} />
            <h3 className="font-medium">Последние импорты</h3>
            <DataTable columns={batchColumns} rows={recentBatches} getRowId={(b) => b.id} />
          </>
        )}
      </CardContent>
    </Card>
  )
}
