import type { GuestProfile } from '@/entities/guest'
import { RecommendationList } from '@/entities/recommendation'
import { useAuth } from '@/entities/user'
import { RefreshRecommendationsButton } from '@/features/refresh-recommendations'
import { Badge } from '@/shared/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { MovementMap } from './movement-map'

export function OverviewTab({ profile }: { profile: GuestProfile }) {
  const { can } = useAuth()

  return (
    <div className="flex flex-col gap-6">
      {can('recommendation.read') && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Следующий шаг</CardTitle>
            {can('recommendation.refresh') && <RefreshRecommendationsButton guestId={profile.id} />}
          </CardHeader>
          <CardContent>
            <RecommendationList guestId={profile.id} />
          </CardContent>
        </Card>
      )}
      {can('guest.event.read') && <MovementMap guestId={profile.id} />}
      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-semibold">Сегменты</h2>
        {profile.segments.length === 0 ? (
          <p className="text-sm text-muted-foreground">Гость не входит в сегменты</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {profile.segments.map((segment) => (
              <Badge key={segment.segmentId} variant="secondary">
                {segment.name}
              </Badge>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
