import { useState } from 'react'
import type { Tenant } from '@/entities/tenant'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteTenant, useUpdateTenant } from '../api/use-tenant-mutations'
import { TenantFormDialog } from './tenant-form-dialog'

type Dialog = 'edit' | 'toggle' | 'delete'

interface TenantActionsProps {
  tenant: Tenant
  onDeleted: () => void
}

export function TenantActions({ tenant, onDeleted }: TenantActionsProps) {
  const [dialog, setDialog] = useState<Dialog | null>(null)
  const update = useUpdateTenant()
  const remove = useDeleteTenant()
  const close = () => setDialog(null)
  const error = update.error ?? remove.error

  return (
    <div className="flex flex-col gap-3">
      {error && <ErrorBanner error={error} />}
      <div className="flex flex-wrap gap-2">
        <Button variant="outline" onClick={() => setDialog('edit')}>
          Изменить
        </Button>
        <Button variant="outline" onClick={() => setDialog('toggle')}>
          {tenant.isActive ? 'Деактивировать' : 'Активировать'}
        </Button>
        <Button variant="outline" onClick={() => setDialog('delete')}>
          Удалить
        </Button>
      </div>
      {dialog === 'edit' && (
        <TenantFormDialog open tenant={tenant} onOpenChange={(open) => !open && close()} />
      )}
      <ConfirmDialog
        open={dialog === 'toggle'}
        onOpenChange={(open) => !open && close()}
        title={tenant.isActive ? 'Деактивировать арендатора?' : 'Активировать арендатора?'}
        description={
          tenant.isActive
            ? `Арендатор «${tenant.name}» перестанет принимать импорт, его источники данных будут отключены. Включить источники обратно придётся вручную.`
            : `Арендатор «${tenant.name}» снова будет доступен. Источники данных нужно включить вручную.`
        }
        confirmLabel={tenant.isActive ? 'Деактивировать' : 'Активировать'}
        destructive={tenant.isActive}
        isPending={update.isPending}
        onConfirm={() =>
          update.mutate(
            { id: tenant.id, body: { isActive: !tenant.isActive } },
            { onSettled: close },
          )
        }
      />
      <ConfirmDialog
        open={dialog === 'delete'}
        onOpenChange={(open) => !open && close()}
        title="Удалить арендатора?"
        description={`Арендатор «${tenant.name}» будет удалён, его источники данных отключены. Если у него есть события, удаление будет отклонено.`}
        confirmLabel="Удалить"
        destructive
        isPending={remove.isPending}
        onConfirm={() =>
          remove.mutate(tenant.id, {
            onSuccess: onDeleted,
            onSettled: close,
          })
        }
      />
    </div>
  )
}
