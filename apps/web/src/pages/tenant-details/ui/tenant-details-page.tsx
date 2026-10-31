import { useNavigate, useParams } from 'react-router'
import { useTenant } from '@/entities/tenant'
import { useAuth } from '@/entities/user'
import { useUrlTab } from '@/shared/lib/hooks/use-url-tab'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { PageHeader } from '@/shared/ui/page-header'
import { Skeleton } from '@/shared/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { TenantOverview } from '@/widgets/tenant-overview'
import { GuestsTab } from './guests-tab'
import { SettingsTab } from './settings-tab'
import { SourcesTab } from './sources-tab'

const TABS = ['overview', 'guests', 'sources', 'settings'] as const

export function TenantDetailsPage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { can } = useAuth()
  const [tab, setTab] = useUrlTab(TABS, 'overview')
  const { data: tenant, error, isLoading } = useTenant(id)

  if (isLoading) return <Skeleton className="h-40 w-full" />
  if (error) return <ErrorBanner error={error} />
  if (!tenant) return <EmptyState title="Арендатор не найден" />

  const canSeeGuests = can('guest.profile.search')
  const canSeeSources = can('data_source.read')

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={tenant.name}
        description={tenant.code}
        breadcrumbs={[{ label: 'Арендаторы', to: '/tenants' }, { label: tenant.name }]}
        actions={<ActiveBadge isActive={tenant.isActive} />}
      />
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="overview">Обзор</TabsTrigger>
          {canSeeGuests && <TabsTrigger value="guests">Гости</TabsTrigger>}
          {canSeeSources && <TabsTrigger value="sources">Источники и импорты</TabsTrigger>}
          <TabsTrigger value="settings">Настройки</TabsTrigger>
        </TabsList>
        <TabsContent value="overview" className="pt-4">
          <TenantOverview tenantId={tenant.id} />
        </TabsContent>
        {canSeeGuests && (
          <TabsContent value="guests" className="pt-4">
            <GuestsTab tenantId={tenant.id} />
          </TabsContent>
        )}
        {canSeeSources && (
          <TabsContent value="sources" className="pt-4">
            <SourcesTab tenantId={tenant.id} />
          </TabsContent>
        )}
        <TabsContent value="settings" className="pt-4">
          <SettingsTab tenant={tenant} onDeleted={() => navigate('/tenants')} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
