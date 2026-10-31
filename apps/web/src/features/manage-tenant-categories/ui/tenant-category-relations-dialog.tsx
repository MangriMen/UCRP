import {
  DEFAULT_RELATION_WEIGHT_PCT,
  relationTypeLabels,
  useTenantCategories,
  useTenantCategoryRelations,
  type TenantCategory,
  type TenantCategoryRelation,
} from '@/entities/tenant-category'
import { DataTable, type Column } from '@/shared/ui/data-table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Skeleton } from '@/shared/ui/skeleton'
import { StatusBadge } from '@/shared/ui/status-badge'
import { TenantCategoryRelationsEditor } from './tenant-category-relations-editor'

interface TenantCategoryRelationsDialogProps {
  category: TenantCategory
  canEdit: boolean
  onOpenChange: (open: boolean) => void
}

export function TenantCategoryRelationsDialog({
  category,
  canEdit,
  onOpenChange,
}: TenantCategoryRelationsDialogProps) {
  const relations = useTenantCategoryRelations(category.id)
  const categories = useTenantCategories()
  const activeIds = new Set(categories.data?.map((c) => c.id))
  const isInactive = (relation: TenantCategoryRelation) =>
    categories.data !== undefined && !activeIds.has(relation.relatedCategoryId)

  const columns: Column<TenantCategoryRelation>[] = [
    {
      id: 'category',
      header: 'Категория',
      cell: (r) => (
        <span className="flex flex-wrap items-center gap-2">
          {r.relatedCategoryName ?? r.relatedCategoryId}
          {isInactive(r) && <StatusBadge tone="neutral">Неактивна</StatusBadge>}
        </span>
      ),
    },
    { id: 'type', header: 'Тип связи', cell: (r) => relationTypeLabels[r.relationType] },
    {
      id: 'weight',
      header: 'Вес, %',
      className: 'text-right',
      cell: (r) =>
        r.relationType === 'adjacent' ? (r.weightPct ?? DEFAULT_RELATION_WEIGHT_PCT) : '—',
    },
  ]

  const renderBody = () => {
    if (relations.error) return <ErrorBanner error={relations.error} />
    if (!relations.data || !categories.data) return <Skeleton className="h-24 w-full" />
    if (canEdit) {
      return (
        <TenantCategoryRelationsEditor
          category={category}
          categories={categories.data}
          relations={relations.data.relations}
          onDone={() => onOpenChange(false)}
        />
      )
    }
    return (
      <DataTable
        columns={columns}
        rows={relations.data.relations}
        getRowId={(r) => `${r.relatedCategoryId}-${r.relationType}`}
        emptyTitle="Связей пока нет"
      />
    )
  }

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Связи категории «{category.name}»</DialogTitle>
          <DialogDescription>
            Смежные категории повышают оценку в рекомендациях, прямые конкуренты понижают её. Связь
            действует в обе стороны.
          </DialogDescription>
        </DialogHeader>
        {renderBody()}
      </DialogContent>
    </Dialog>
  )
}
