import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { useSegmentRecommendations } from '../api/use-recommendations'
import type { SegmentRecommendation } from '../model/types'

const columns: Column<SegmentRecommendation>[] = [
  { id: 'tenant', header: 'Арендатор', cell: (t) => t.tenantName },
  { id: 'category', header: 'Категория', cell: (t) => t.categoryId ?? '—' },
  { id: 'guests', header: 'Гостей', cell: (t) => t.guests ?? '—', className: 'text-right' },
]

interface SegmentTopTenantsProps {
  segmentId: string
  isSegmentComputing: boolean
}

export function SegmentTopTenants({ segmentId, isSegmentComputing }: SegmentTopTenantsProps) {
  const { data, error, isLoading } = useSegmentRecommendations(segmentId, isSegmentComputing)

  return (
    <Card>
      <CardHeader>
        <CardTitle>Кого приглашать</CardTitle>
      </CardHeader>
      <CardContent>
        <DataTable
          columns={columns}
          rows={data?.items}
          getRowId={(t) => t.tenantId}
          isLoading={isLoading}
          error={error}
          emptyTitle="Рекомендаций для сегмента пока нет"
        />
      </CardContent>
    </Card>
  )
}
