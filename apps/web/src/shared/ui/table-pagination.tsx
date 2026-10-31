import type { Schema } from '@/shared/api/types'
import { PAGE_SIZE_OPTIONS } from '@/shared/lib/list-params'
import { Button } from './button'
import { OptionSelect } from './option-select'

interface TablePaginationProps {
  meta: Schema<'PageMeta'>
  onPageChange: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
}

const pageSizeItems = PAGE_SIZE_OPTIONS.map((size) => ({
  value: String(size),
  label: `${size} на странице`,
}))

export function TablePagination({ meta, onPageChange, onPageSizeChange }: TablePaginationProps) {
  const { page, totalPages, total, pageSize } = meta
  return (
    <div className="flex flex-wrap items-center justify-between gap-4 pt-4 text-sm">
      <span className="text-muted-foreground">Всего: {total}</span>
      <div className="flex items-center gap-2">
        {onPageSizeChange && (
          <OptionSelect
            items={pageSizeItems}
            value={String(pageSize)}
            onValueChange={(value) => onPageSizeChange(Number(value))}
            ariaLabel="Размер страницы"
            className="w-44"
          />
        )}
        <Button
          variant="outline"
          size="sm"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          Назад
        </Button>
        <span>
          {page} / {Math.max(totalPages, 1)}
        </span>
        <Button
          variant="outline"
          size="sm"
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          Вперёд
        </Button>
      </div>
    </div>
  )
}
