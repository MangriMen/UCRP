import type { ReactNode } from 'react'
import { useMalls } from '@/entities/mall'
import type { Tenant } from '@/entities/tenant'
import { useTenantCategories } from '@/entities/tenant-category'
import { useAuth } from '@/entities/user'
import { TenantActions } from '@/features/manage-tenants'
import { formatDateTime } from '@/shared/lib/format'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Card, CardContent } from '@/shared/ui/card'

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1 sm:flex-row sm:gap-4">
      <dt className="w-48 shrink-0 text-sm text-muted-foreground">{label}</dt>
      <dd>{children}</dd>
    </div>
  )
}

interface SettingsTabProps {
  tenant: Tenant
  onDeleted: () => void
}

export function SettingsTab({ tenant, onDeleted }: SettingsTabProps) {
  const { can } = useAuth()
  const categories = useTenantCategories()
  const malls = useMalls()
  const category = categories.data?.find((item) => item.id === tenant.categoryId)
  const mall = malls.data?.find((item) => item.id === tenant.mallId)

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardContent>
          <dl className="flex flex-col gap-3">
            <Row label="Код">{tenant.code}</Row>
            <Row label="Название">{tenant.name}</Row>
            <Row label="Категория">{category?.name ?? '—'}</Row>
            <Row label="ТЦ">{mall?.name ?? '—'}</Row>
            <Row label="Этаж">{tenant.floor ?? '—'}</Row>
            <Row label="Приоритет">{tenant.priority}</Row>
            <Row label="Статус">
              <ActiveBadge isActive={tenant.isActive} />
            </Row>
            <Row label="Создан">{formatDateTime(tenant.createdAt)}</Row>
            <Row label="Изменён">{formatDateTime(tenant.updatedAt)}</Row>
          </dl>
        </CardContent>
      </Card>
      {can('reference.write') && <TenantActions tenant={tenant} onDeleted={onDeleted} />}
    </div>
  )
}
