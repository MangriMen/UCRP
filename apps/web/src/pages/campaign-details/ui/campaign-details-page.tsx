import { Link, useParams } from 'react-router'
import { CampaignResults, CampaignStatusBadge, useCampaign } from '@/entities/campaign'
import { useAuth } from '@/entities/user'
import { formatDate } from '@/shared/lib/format'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'

function Fact({ title, value }: { title: string; value: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-normal text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent className="text-lg font-semibold">{value}</CardContent>
    </Card>
  )
}

export function CampaignDetailsPage() {
  const { id = '' } = useParams()
  const { role } = useAuth()
  const { data: campaign, error, isLoading } = useCampaign(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!campaign) return <EmptyState title="Кампания не найдена" />

  const canSeeResults = role === 'marketer' || role === 'manager'

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={campaign.name}
        actions={
          <>
            <CampaignStatusBadge status={campaign.status} />
            <Button variant="outline" render={<Link to="/campaigns" />}>
              К списку
            </Button>
          </>
        }
      />
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Fact title="Арендатор" value={campaign.targetTenantName ?? '—'} />
        <Fact title="Сегмент" value={campaign.segmentName ?? '—'} />
        <Fact title="Окно атрибуции" value={`${campaign.attributionDays} дн.`} />
        <Fact
          title="Период"
          value={
            campaign.startsAt
              ? `${formatDate(campaign.startsAt)} — ${campaign.endsAt ? formatDate(campaign.endsAt) : '…'}`
              : '—'
          }
        />
      </div>
      {canSeeResults && (
        <section className="flex flex-col gap-3">
          <h2 className="text-lg font-semibold">Результаты</h2>
          <CampaignResults campaignId={campaign.id} />
        </section>
      )}
    </div>
  )
}
