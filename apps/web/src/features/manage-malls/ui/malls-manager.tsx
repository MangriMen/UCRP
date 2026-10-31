import { useState } from 'react'
import { useMalls, type Mall } from '@/entities/mall'
import { ActiveBadge } from '@/shared/ui/active-badge'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteMall } from '../api/use-mall-mutations'
import { MallFormDialog } from './mall-form-dialog'

interface MallsManagerProps {
  canEdit: boolean
}

interface FormState {
  mall: Mall | null
}

export function MallsManager({ canEdit }: MallsManagerProps) {
  const malls = useMalls()
  const remove = useDeleteMall()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<Mall | null>(null)

  const columns: Column<Mall>[] = [
    { id: 'code', header: 'Код', cell: (m) => m.code },
    { id: 'name', header: 'Название', cell: (m) => m.name },
    {
      id: 'floors',
      header: 'Этажи',
      cell: (m) => (m.floors.length > 0 ? m.floors.join(', ') : '—'),
    },
    { id: 'status', header: 'Активен', cell: (m) => <ActiveBadge isActive={m.isActive} /> },
    ...(canEdit
      ? [
          {
            id: 'actions',
            header: '',
            className: 'text-right',
            cell: (m: Mall) => (
              <div className="flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={() => setForm({ mall: m })}>
                  Изменить
                </Button>
                <Button variant="outline" size="sm" onClick={() => setDeleting(m)}>
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
          <Button onClick={() => setForm({ mall: null })}>Добавить ТЦ</Button>
        </div>
      )}
      {remove.error && <ErrorBanner error={remove.error} />}
      <DataTable
        columns={columns}
        rows={malls.data}
        getRowId={(m) => m.id}
        isLoading={malls.isLoading}
        error={malls.error}
        emptyTitle="Торговых центров пока нет"
      />
      {form && (
        <MallFormDialog open mall={form.mall} onOpenChange={(open) => !open && setForm(null)} />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Удалить ТЦ?"
        description={`ТЦ «${deleting?.name ?? ''}» будет удалён. Если в нём есть арендаторы, удаление будет отклонено.`}
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
