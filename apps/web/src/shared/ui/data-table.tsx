import type { ReactNode } from 'react'
import { ArrowDownIcon, ArrowUpDownIcon, ArrowUpIcon } from 'lucide-react'
import type { Schema } from '@/shared/api/types'
import { nextSort, type SortState } from '@/shared/lib/list-params'
import { Button } from './button'
import { EmptyState } from './empty-state'
import { ErrorBanner } from './error-banner'
import { Skeleton } from './skeleton'
import { TablePagination } from './table-pagination'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './table'

export interface Column<T> {
  id: string
  header: string
  cell: (row: T) => ReactNode
  className?: string
  sortable?: boolean
}

interface DataTableProps<T> {
  columns: Column<T>[]
  rows: T[] | undefined
  getRowId: (row: T) => string
  isLoading?: boolean
  error?: unknown
  emptyTitle?: string
  onResetFilters?: () => void
  meta?: Schema<'PageMeta'>
  onPageChange?: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
  sort?: SortState
  onSortChange?: (sort: SortState | undefined) => void
  onRowClick?: (row: T) => void
}

const SKELETON_ROWS = 5

const ariaSort = (sort: SortState | undefined, id: string) => {
  if (sort?.field !== id) return 'none'
  return sort.direction === 'asc' ? 'ascending' : 'descending'
}

function SortIcon({ sort, id }: { sort: SortState | undefined; id: string }) {
  if (sort?.field !== id) return <ArrowUpDownIcon className="size-3.5 opacity-50" />
  return sort.direction === 'asc' ? (
    <ArrowUpIcon className="size-3.5" />
  ) : (
    <ArrowDownIcon className="size-3.5" />
  )
}

export function DataTable<T>({
  columns,
  rows,
  getRowId,
  isLoading,
  error,
  emptyTitle = 'Нет данных',
  onResetFilters,
  meta,
  onPageChange,
  onPageSizeChange,
  sort,
  onSortChange,
  onRowClick,
}: DataTableProps<T>) {
  if (error) return <ErrorBanner error={error} />
  if (!isLoading && rows?.length === 0) {
    const resetAction = onResetFilters && (
      <Button type="button" variant="outline" onClick={onResetFilters}>
        Сбросить фильтры
      </Button>
    )
    return <EmptyState title={emptyTitle} action={resetAction} />
  }

  return (
    <div>
      <Table>
        <TableHeader className="sticky top-0 z-10 bg-background">
          <TableRow>
            {columns.map((c) => (
              <TableHead
                key={c.id}
                className={c.className}
                aria-sort={c.sortable && onSortChange ? ariaSort(sort, c.id) : undefined}
              >
                {c.sortable && onSortChange ? (
                  <button
                    type="button"
                    className="inline-flex items-center gap-1 font-medium hover:text-foreground"
                    onClick={() => onSortChange(nextSort(sort, c.id))}
                  >
                    {c.header}
                    <SortIcon sort={sort} id={c.id} />
                  </button>
                ) : (
                  c.header
                )}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading
            ? Array.from({ length: SKELETON_ROWS }, (_, i) => (
                <TableRow key={i}>
                  {columns.map((c) => (
                    <TableCell key={c.id}>
                      <Skeleton className="h-4 w-full" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            : rows?.map((row) => (
                <TableRow
                  key={getRowId(row)}
                  className={onRowClick ? 'cursor-pointer' : undefined}
                  onClick={onRowClick && (() => onRowClick(row))}
                >
                  {columns.map((c) => (
                    <TableCell key={c.id} className={c.className}>
                      {c.cell(row)}
                    </TableCell>
                  ))}
                </TableRow>
              ))}
        </TableBody>
      </Table>
      {meta && onPageChange && (
        <TablePagination
          meta={meta}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
        />
      )}
    </div>
  )
}
