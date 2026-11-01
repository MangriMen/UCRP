import { formatNumber, formatSuppressed, SUPPRESSED_LABEL } from '@/shared/lib/format'
import { Alert, AlertDescription } from '@/shared/ui/alert'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { OptionSelect } from '@/shared/ui/option-select'
import { useSegmentRecommendations } from '../api/use-recommendations'
import {
  SEGMENT_TOP_TENANTS_LIMITS,
  SEGMENT_TOP_TENANTS_SUPPRESSION_THRESHOLD,
} from '../model/keys'
import type { SegmentRecommendation } from '../model/types'

const limitOptions = SEGMENT_TOP_TENANTS_LIMITS.map((limit) => ({
  value: String(limit),
  label: `Топ ${limit}`,
}))

const count = (value: number | null | undefined) => formatSuppressed(value, formatNumber)

const reach = (tenant: SegmentRecommendation) =>
  tenant.guests == null || tenant.eligibleCount == null
    ? SUPPRESSED_LABEL
    : `${formatNumber(tenant.guests)} из ${formatNumber(tenant.eligibleCount)}`

const averageScore = (tenant: SegmentRecommendation) =>
  formatSuppressed(tenant.averageScore, (score) => score.toFixed(2))

interface SegmentTopTenantsProps {
  segmentId: string
  isSegmentComputing: boolean
  limit: number
  onLimitChange: (limit: number) => void
  categoryName: (categoryId: string) => string | undefined
}

export function SegmentTopTenants({
  segmentId,
  isSegmentComputing,
  limit,
  onLimitChange,
  categoryName,
}: SegmentTopTenantsProps) {
  const { data, error, isLoading } = useSegmentRecommendations(segmentId, isSegmentComputing, limit)
  const hasSuppressed = data?.items.some((tenant) => tenant.suppressed) ?? false

  const columns: Column<SegmentRecommendation>[] = [
    { id: 'tenant', header: 'Арендатор', cell: (t) => t.tenantName },
    {
      id: 'category',
      header: 'Категория',
      cell: (t) => (t.categoryId ? (categoryName(t.categoryId) ?? t.categoryId) : '—'),
    },
    { id: 'reach', header: 'Рекомендован гостям', cell: reach, className: 'text-right' },
    { id: 'score', header: 'Средняя оценка', cell: averageScore, className: 'text-right' },
    {
      id: 'incomplete',
      header: 'Без свежего расчёта',
      cell: (t) => count(t.incompleteCount),
      className: 'text-right',
    },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Кого приглашать</CardTitle>
        <CardAction>
          <OptionSelect
            items={limitOptions}
            value={String(limit)}
            onValueChange={(next) => onLimitChange(Number(next))}
            ariaLabel="Сколько арендаторов показать"
          />
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {hasSuppressed && (
          <Alert>
            <AlertDescription>
              Часть значений скрыта: группы меньше {SEGMENT_TOP_TENANTS_SUPPRESSION_THRESHOLD}{' '}
              гостей не показываются, чтобы нельзя было определить отдельного гостя.
            </AlertDescription>
          </Alert>
        )}
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
