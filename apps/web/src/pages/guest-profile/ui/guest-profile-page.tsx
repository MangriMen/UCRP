import { Link, useNavigate, useParams } from 'react-router'
import { useGuest } from '@/entities/guest'
import { RecommendationList } from '@/entities/recommendation'
import { useAuth } from '@/entities/user'
import { MergeGuestsDialog } from '@/features/merge-guests'
import { RefreshRecommendationsButton } from '@/features/refresh-recommendations'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { ContactHistory } from './contact-history'
import { EventFeed } from './event-feed'
import { MovementMap } from './movement-map'
import { ProfileCard } from './profile-card'

export function GuestProfilePage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { role } = useAuth()
  const { data: profile, error, isLoading } = useGuest(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!profile) return <EmptyState title="Гость не найден" />

  const canSeeRecommendations = role === 'marketer' || role === 'analyst'
  const canSeeMovement = role === 'marketer' || role === 'manager' || role === 'analyst'

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Профиль гостя"
        actions={
          <>
            {role === 'analyst' && profile.status === 'active' && (
              <MergeGuestsDialog
                guestId={profile.id}
                onMerged={(result) => navigate(`/guests/${result.winnerGuestId}`)}
              />
            )}
            <Button variant="outline" render={<Link to="/guests" />}>
              К поиску
            </Button>
          </>
        }
      />
      <ProfileCard profile={profile} />
      {canSeeRecommendations && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Следующий шаг</CardTitle>
            <RefreshRecommendationsButton guestId={profile.id} />
          </CardHeader>
          <CardContent>
            <RecommendationList guestId={profile.id} />
          </CardContent>
        </Card>
      )}
      {canSeeMovement && <MovementMap guestId={profile.id} />}
      <EventFeed guestId={profile.id} />
      <ContactHistory guestId={profile.id} />
    </div>
  )
}
