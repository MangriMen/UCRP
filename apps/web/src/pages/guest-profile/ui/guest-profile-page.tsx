import { useNavigate, useParams } from 'react-router'
import { GuestStatusBadge, useGuest } from '@/entities/guest'
import { useAuth } from '@/entities/user'
import { DeleteGuestButton } from '@/features/delete-guest'
import { MergeGuestsDialog } from '@/features/merge-guests'
import { formatDateTime, formatMoney, formatNumber } from '@/shared/lib/format'
import { useUrlTab } from '@/shared/lib/hooks/use-url-tab'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { StatCard, StatGrid } from '@/shared/ui/stat-card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { EventsTable } from './events-table'
import { OverviewTab } from './overview-tab'
import { ProfileTab } from './profile-tab'
import { TenantsTab } from './tenants-tab'

const TABS = ['overview', 'history', 'purchases', 'tenants', 'profile'] as const

export function GuestProfilePage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { can } = useAuth()
  const [tab, setTab] = useUrlTab(TABS, 'overview')
  const { data: profile, error, isLoading } = useGuest(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!profile) return <EmptyState title="Гость не найден" />

  const title = profile.displayName ?? 'Без имени'
  const isActive = profile.status === 'active'
  const canSeeEvents = can('guest.event.read')
  const canSeeAmount = can('field.event.amount.read')

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={title}
        breadcrumbs={[{ label: 'Гости', to: '/guests' }, { label: title }]}
        actions={
          <>
            <GuestStatusBadge status={profile.status} />
            {can('identity.decide') && isActive && (
              <MergeGuestsDialog
                guestId={profile.id}
                onMerged={(result) => navigate(`/guests/${result.winnerGuestId}`)}
              />
            )}
            {can('guest.profile.delete') && isActive && (
              <DeleteGuestButton guestId={profile.id} onDeleted={() => navigate('/guests')} />
            )}
          </>
        }
      />
      <StatGrid>
        <StatCard label="Событий" value={formatNumber(profile.eventCount ?? 0)} />
        <StatCard label="Покупок" value={formatNumber(profile.purchaseCount ?? 0)} />
        {canSeeAmount && profile.totalAmountMinor != null && (
          <StatCard label="Сумма покупок" value={formatMoney(profile.totalAmountMinor)} />
        )}
        <StatCard label="Арендаторов" value={formatNumber(profile.tenantCount)} />
        <StatCard
          label="Последняя активность"
          value={profile.lastSeenAt ? formatDateTime(profile.lastSeenAt) : '—'}
          hint={
            profile.firstSeenAt
              ? `Первое появление: ${formatDateTime(profile.firstSeenAt)}`
              : undefined
          }
        />
      </StatGrid>
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="overview">Обзор</TabsTrigger>
          {canSeeEvents && <TabsTrigger value="history">История</TabsTrigger>}
          {canSeeEvents && <TabsTrigger value="purchases">Покупки</TabsTrigger>}
          <TabsTrigger value="tenants">Арендаторы</TabsTrigger>
          <TabsTrigger value="profile">Профиль</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" className="pt-4">
          <OverviewTab profile={profile} />
        </TabsContent>
        {canSeeEvents && (
          <TabsContent value="history" className="pt-4">
            <EventsTable guestId={profile.id} tenants={profile.tenants} />
          </TabsContent>
        )}
        {canSeeEvents && (
          <TabsContent value="purchases" className="pt-4">
            <EventsTable guestId={profile.id} tenants={profile.tenants} purchasesOnly />
          </TabsContent>
        )}
        <TabsContent value="tenants" className="pt-4">
          <TenantsTab profile={profile} />
        </TabsContent>
        <TabsContent value="profile" className="pt-4">
          <ProfileTab profile={profile} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
