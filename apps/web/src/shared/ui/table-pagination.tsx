import type { Schema } from '@/shared/api/types'
import { Button } from './button'

interface TablePaginationProps {
  meta: Schema<'PageMeta'>
  onPageChange: (page: number) => void
}

export function TablePagination({ meta, onPageChange }: TablePaginationProps) {
  const { page, totalPages, total } = meta
  return (
    <div className="flex items-center justify-between gap-4 pt-4 text-sm">
      <span className="text-muted-foreground">Всего: {total}</span>
      <div className="flex items-center gap-2">
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
