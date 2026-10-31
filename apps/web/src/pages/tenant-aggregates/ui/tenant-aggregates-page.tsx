import { useAuth } from '@/entities/user'
import { TenantOverview } from '@/widgets/tenant-overview'
import { EmptyState } from '@/shared/ui/empty-state'
import { PageHeader } from '@/shared/ui/page-header'

export function TenantAggregatesPage() {
  const { user } = useAuth()
  const tenantId = user?.tenantId ?? null

  if (!tenantId) {
    return <EmptyState title="Арендатор не привязан к вашей учётной записи" />
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Мой арендатор"
        description="Агрегированная статистика без данных об отдельных гостях"
      />
      <TenantOverview tenantId={tenantId} />
    </div>
  )
}
