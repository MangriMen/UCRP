import type { ReactNode } from 'react'
import type { Schema } from '@/shared/api/types'
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
}

interface DataTableProps<T> {
  columns: Column<T>[]
  rows: T[] | undefined
  getRowId: (row: T) => string
  isLoading?: boolean
  error?: unknown
  emptyTitle?: string
  meta?: Schema<'PageMeta'>
  onPageChange?: (page: number) => void
  onRowClick?: (row: T) => void
}

const SKELETON_ROWS = 5

export function DataTable<T>({
  columns,
  rows,
  getRowId,
  isLoading,
  error,
  emptyTitle = 'Нет данных',
  meta,
  onPageChange,
  onRowClick,
}: DataTableProps<T>) {
  if (error) return <ErrorBanner error={error} />
  if (!isLoading && rows?.length === 0) return <EmptyState title={emptyTitle} />

  return (
    <div>
      <Table>
        <TableHeader>
          <TableRow>
            {columns.map((c) => (
              <TableHead key={c.id} className={c.className}>
                {c.header}
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
      {meta && onPageChange && <TablePagination meta={meta} onPageChange={onPageChange} />}
    </div>
  )
}
