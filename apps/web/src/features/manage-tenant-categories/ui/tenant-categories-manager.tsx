import { useState } from 'react'
import { useTenantCategories, type TenantCategory } from '@/entities/tenant-category'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { useDeleteTenantCategory } from '../api/use-tenant-category-mutations'
import { TenantCategoryFormDialog } from './tenant-category-form-dialog'
import { TenantCategoryRelationsDialog } from './tenant-category-relations-dialog'

interface TenantCategoriesManagerProps {
  canEdit: boolean
  relationsCategoryId: string | null
  onRelationsCategoryChange: (categoryId: string | null) => void
}

interface FormState {
  category: TenantCategory | null
}

export function TenantCategoriesManager({
  canEdit,
  relationsCategoryId,
  onRelationsCategoryChange,
}: TenantCategoriesManagerProps) {
  const categories = useTenantCategories()
  const remove = useDeleteTenantCategory()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleting, setDeleting] = useState<TenantCategory | null>(null)
  const relationsCategory = categories.data?.find((c) => c.id === relationsCategoryId)

  const columns: Column<TenantCategory>[] = [
    { id: 'code', header: 'Код', cell: (c) => c.code },
    { id: 'name', header: 'Название', cell: (c) => c.name },
    { id: 'description', header: 'Описание', cell: (c) => c.description ?? '—' },
    {
      id: 'actions',
      header: '',
      className: 'text-right',
      cell: (c) => (
        <div className="flex flex-wrap justify-end gap-2">
          <Button variant="outline" size="sm" onClick={() => onRelationsCategoryChange(c.id)}>
            Связи
          </Button>
          {canEdit && (
            <>
              <Button variant="outline" size="sm" onClick={() => setForm({ category: c })}>
                Изменить
              </Button>
              <Button variant="outline" size="sm" onClick={() => setDeleting(c)}>
                Удалить
              </Button>
            </>
          )}
        </div>
      ),
    },
  ]

  return (
    <div className="flex flex-col gap-4">
      {canEdit && (
        <div className="flex justify-end">
          <Button onClick={() => setForm({ category: null })}>Добавить категорию</Button>
        </div>
      )}
      {remove.error && <ErrorBanner error={remove.error} />}
      <DataTable
        columns={columns}
        rows={categories.data}
        getRowId={(c) => c.id}
        isLoading={categories.isLoading}
        error={categories.error}
        emptyTitle="Категорий пока нет"
      />
      {form && (
        <TenantCategoryFormDialog
          open
          category={form.category}
          onOpenChange={(open) => !open && setForm(null)}
        />
      )}
      {relationsCategory && (
        <TenantCategoryRelationsDialog
          key={relationsCategory.id}
          category={relationsCategory}
          canEdit={canEdit}
          onOpenChange={(open) => !open && onRelationsCategoryChange(null)}
        />
      )}
      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Удалить категорию?"
        description={`Категория «${deleting?.name ?? ''}» будет удалена. Если она назначена арендаторам или связана с другими категориями, удаление будет отклонено.`}
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
