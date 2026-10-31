import { useState } from 'react'
import { useTenants, type Tenant } from '@/entities/tenant'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteTenant } from '../api/use-tenant-mutations'
import { TenantFormDialog } from './tenant-form-dialog'

interface TenantsManagerProps {
  canEdit: boolean
}

interface FormState {
  tenant: Tenant | null
}

export function TenantsManager({ canEdit }: TenantsManagerProps) {
  const [page, setPage] = useState(1)
  const tenants = useTenants({ page })
  const remove = useDeleteTenant()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<Tenant | null>(null)

  const columns: Column<Tenant>[] = [
    { id: 'code', header: 'Код', cell: (t) => t.code },
    { id: 'name', header: 'Название', cell: (t) => t.name },
    { id: 'category', header: 'Категория', cell: (t) => t.categoryName ?? '—' },
    { id: 'floor', header: 'Этаж', cell: (t) => t.floor, className: 'text-right' },
    { id: 'priority', header: 'Приоритет', cell: (t) => t.priority, className: 'text-right' },
    { id: 'status', header: 'Статус', cell: (t) => <ActiveBadge isActive={t.isActive} /> },
    ...(canEdit
      ? [
          {
            id: 'actions',
            header: '',
            className: 'text-right',
            cell: (t: Tenant) => (
              <div className="flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={() => setForm({ tenant: t })}>
                  Изменить
                </Button>
                <Button variant="outline" size="sm" onClick={() => setDeleting(t)}>
                  Удалить
                </Button>
              </div>
            ),
          },
        ]
      : []),
  ]

  return (
    <div className="flex flex-col gap-4">
      {canEdit && (
        <div className="flex justify-end">
          <Button onClick={() => setForm({ tenant: null })}>Добавить арендатора</Button>
        </div>
      )}
      {remove.error && <ErrorBanner error={remove.error} />}
      <DataTable
        columns={columns}
        rows={tenants.data?.items}
        getRowId={(t) => t.id}
        isLoading={tenants.isLoading}
        error={tenants.error}
        emptyTitle="Арендаторов пока нет"
        meta={tenants.data?.meta}
        onPageChange={setPage}
      />
      {form && (
        <TenantFormDialog
          open
          tenant={form.tenant}
          onOpenChange={(open) => !open && setForm(null)}
        />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Удалить арендатора?"
        description={`Арендатор «${deleting?.name ?? ''}» будет деактивирован. Если у него есть события, удаление будет отклонено.`}
        confirmLabel="Удалить"
        destructive
        isPending={remove.isPending}
        onConfirm={() => {
          if (deleting) remove.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
      />
    </div>
  )
}
