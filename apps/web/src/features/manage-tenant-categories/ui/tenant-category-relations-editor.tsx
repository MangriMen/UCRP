import { useRef, useState } from 'react'
import { Trash2Icon } from 'lucide-react'
import {
  DEFAULT_RELATION_WEIGHT_PCT,
  relationTypeLabels,
  type CategoryRelationType,
  type TenantCategory,
  type TenantCategoryRelation,
} from '@/entities/tenant-category'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { EmptyState } from '@/shared/ui/empty-state'
import { ErrorBanner } from '@/shared/ui/error-banner'
import { Input } from '@/shared/ui/input'
import { OptionSelect, type Option } from '@/shared/ui/option-select'
import { useReplaceTenantCategoryRelations } from '../api/use-tenant-category-mutations'

interface DraftRow {
  key: number
  relatedCategoryId: string
  relationType: CategoryRelationType
  weightPct: string
}

interface TenantCategoryRelationsEditorProps {
  category: TenantCategory
  categories: TenantCategory[]
  relations: TenantCategoryRelation[]
  onDone: () => void
}

const relationTypeOptions: Option[] = Object.entries(relationTypeLabels).map(([value, label]) => ({
  value,
  label,
}))

const isValidWeight = (value: string) => {
  const weight = Number(value)
  return value.trim() !== '' && Number.isInteger(weight) && weight >= 0 && weight <= 100
}

const isRowValid = (row: DraftRow) =>
  row.relationType === 'competitor' || isValidWeight(row.weightPct)

export function TenantCategoryRelationsEditor({
  category,
  categories,
  relations,
  onDone,
}: TenantCategoryRelationsEditorProps) {
  const replace = useReplaceTenantCategoryRelations(category.id)
  const nextKey = useRef(relations.length)
  const [rows, setRows] = useState<DraftRow[]>(() =>
    relations.map((r, index) => ({
      key: index,
      relatedCategoryId: r.relatedCategoryId,
      relationType: r.relationType,
      weightPct: String(r.weightPct ?? DEFAULT_RELATION_WEIGHT_PCT),
    })),
  )

  const candidates = categories.filter((c) => c.id !== category.id)
  const isTaken = (categoryId: string, exceptKey?: number) =>
    rows.some((r) => r.relatedCategoryId === categoryId && r.key !== exceptKey)
  const available = candidates.filter((c) => !isTaken(c.id))

  const nameOf = (categoryId: string) =>
    candidates.find((c) => c.id === categoryId)?.name ??
    relations.find((r) => r.relatedCategoryId === categoryId)?.relatedCategoryName ??
    categoryId

  const optionsFor = (row: DraftRow): Option[] => {
    const options = candidates
      .filter((c) => !isTaken(c.id, row.key))
      .map((c) => ({ value: c.id, label: c.name }))
    if (!options.some((o) => o.value === row.relatedCategoryId)) {
      options.unshift({
        value: row.relatedCategoryId,
        label: `${nameOf(row.relatedCategoryId)} (неактивна)`,
      })
    }
    return options
  }

  const updateRow = (key: number, patch: Partial<DraftRow>) =>
    setRows((current) => current.map((r) => (r.key === key ? { ...r, ...patch } : r)))

  const addRow = () => {
    const [first] = available
    if (!first) return
    const key = nextKey.current++
    setRows((current) => [
      ...current,
      {
        key,
        relatedCategoryId: first.id,
        relationType: 'adjacent',
        weightPct: String(DEFAULT_RELATION_WEIGHT_PCT),
      },
    ])
  }

  const save = () =>
    replace.mutate(
      rows.map((r) => ({
        categoryId: category.id,
        relatedCategoryId: r.relatedCategoryId,
        relationType: r.relationType,
        weightPct: r.relationType === 'adjacent' ? Number(r.weightPct) : null,
      })),
      { onSuccess: onDone },
    )

  return (
    <div className="flex flex-col gap-4">
      {replace.error && <ErrorBanner error={replace.error} />}
      {rows.length === 0 ? (
        <EmptyState
          title="Связей пока нет"
          description="Добавьте смежную категорию или конкурента"
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {rows.map((row, index) => {
            const position = index + 1
            const weightInvalid = row.relationType === 'adjacent' && !isValidWeight(row.weightPct)
            return (
              <li key={row.key} className="flex flex-wrap items-start gap-2">
                <OptionSelect
                  ariaLabel={`Категория, связь ${position}`}
                  items={optionsFor(row)}
                  value={row.relatedCategoryId}
                  onValueChange={(relatedCategoryId) => updateRow(row.key, { relatedCategoryId })}
                  className="min-w-40 flex-1"
                />
                <OptionSelect
                  ariaLabel={`Тип, связь ${position}`}
                  items={relationTypeOptions}
                  value={row.relationType}
                  onValueChange={(relationType) =>
                    updateRow(row.key, { relationType: relationType as CategoryRelationType })
                  }
                  className="w-44"
                />
                {row.relationType === 'adjacent' && (
                  <div className="flex flex-col gap-1">
                    <Input
                      type="number"
                      min={0}
                      max={100}
                      step={1}
                      aria-label={`Вес в процентах, связь ${position}`}
                      aria-invalid={weightInvalid}
                      value={row.weightPct}
                      onChange={(event) => updateRow(row.key, { weightPct: event.target.value })}
                      className="w-24"
                    />
                    {weightInvalid && <p className="text-xs text-destructive">Вес от 0 до 100</p>}
                  </div>
                )}
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  aria-label={`Удалить связь с «${nameOf(row.relatedCategoryId)}»`}
                  onClick={() => setRows((current) => current.filter((r) => r.key !== row.key))}
                >
                  <Trash2Icon />
                </Button>
              </li>
            )
          })}
        </ul>
      )}
      <div>
        <Button type="button" variant="outline" onClick={addRow} disabled={available.length === 0}>
          Добавить связь
        </Button>
      </div>
      <DialogFooter>
        <Button type="button" variant="outline" onClick={onDone}>
          Отмена
        </Button>
        <Button
          type="button"
          onClick={save}
          disabled={replace.isPending || !rows.every(isRowValid)}
        >
          Сохранить
        </Button>
      </DialogFooter>
    </div>
  )
}
