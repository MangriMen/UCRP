import { Link, useNavigate, useParams } from 'react-router'
import { useGuest } from '@/entities/guest'
import { useAuth } from '@/entities/user'
import { MergeGuestsDialog } from '@/features/merge-guests'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { ContactHistory } from './contact-history'
import { EventFeed } from './event-feed'
import { ProfileCard } from './profile-card'

export function GuestProfilePage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { role } = useAuth()
  const { data: profile, error, isLoading } = useGuest(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!profile) return <EmptyState title="Гость не найден" />

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
      <Card>
        <CardHeader>
          <CardTitle>Следующий шаг</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Персональные рекомендации появятся здесь
        </CardContent>
      </Card>
      <EventFeed guestId={profile.id} />
      <ContactHistory guestId={profile.id} />
    </div>
  )
}
