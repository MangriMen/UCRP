import { formatDateTime } from '@/shared/lib/format'
import { Badge } from '@/shared/ui/badge'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Skeleton } from '@/shared/ui/skeleton'
import { useGuestRecommendations } from '../api/use-recommendations'
import { isLowData } from '../model/factors'
import { WhyPopover } from './why-popover'

export function RecommendationList({ guestId }: { guestId: string }) {
  const { data, error, isLoading } = useGuestRecommendations(guestId)

  if (isLoading) return <Skeleton className="h-24 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!data || data.items.length === 0) {
    return (
      <EmptyState
        title="Рекомендаций пока нет"
        description="Они появятся после расчёта: автоматически по новым событиям или по кнопке «Обновить»."
      />
    )
  }

  return (
    <div className="flex flex-col gap-3">
      <ol className="flex flex-col gap-2">
        {[...data.items]
          .sort((a, b) => a.rank - b.rank)
          .map((item) => (
            <li
              key={item.id}
              className="flex flex-wrap items-center justify-between gap-2 rounded-lg border p-3"
            >
              <div className="flex flex-col">
                <span className="flex flex-wrap items-center gap-2 font-medium">
                  {item.rank}. {item.tenantName ?? item.tenantId}
                  {isLowData(item) && <Badge variant="outline">Мало данных</Badge>}
                </span>
                <span className="text-sm text-muted-foreground">
                  {[item.categoryName, item.floor != null ? `${item.floor} этаж` : null]
                    .filter(Boolean)
                    .join(' · ') || '—'}
                </span>
              </div>
              <WhyPopover recommendation={item} />
            </li>
          ))}
      </ol>
      {data.computedAt && (
        <p className="text-xs text-muted-foreground">
          Рассчитано: {formatDateTime(data.computedAt)}
        </p>
      )}
    </div>
  )
}
