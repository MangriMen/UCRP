import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { DataTable, type Column } from './data-table'

interface Row {
  id: string
  name: string
}

const columns: Column<Row>[] = [{ id: 'name', header: 'Имя', cell: (r) => r.name }]
const getRowId = (r: Row) => r.id

describe('DataTable', () => {
  it('renders rows and handles paging', async () => {
    const onPageChange = vi.fn()
    render(
      <DataTable
        columns={columns}
        rows={[{ id: '1', name: 'Анна' }]}
        getRowId={getRowId}
        meta={{ page: 1, pageSize: 50, total: 120, totalPages: 3 }}
        onPageChange={onPageChange}
      />,
    )
    expect(screen.getByText('Анна')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Вперёд' }))
    expect(onPageChange).toHaveBeenCalledWith(2)
    expect(screen.getByRole('button', { name: 'Назад' })).toBeDisabled()
  })

  it('sorts by a sortable column and reports page size changes', async () => {
    const onSortChange = vi.fn()
    const onPageSizeChange = vi.fn()
    render(
      <DataTable
        columns={[{ ...columns[0], sortable: true }]}
        rows={[{ id: '1', name: 'Анна' }]}
        getRowId={getRowId}
        sort={{ field: 'name', direction: 'asc' }}
        onSortChange={onSortChange}
        meta={{ page: 1, pageSize: 50, total: 1, totalPages: 1 }}
        onPageChange={vi.fn()}
        onPageSizeChange={onPageSizeChange}
      />,
    )
    expect(screen.getByRole('columnheader', { name: /Имя/ })).toHaveAttribute(
      'aria-sort',
      'ascending',
    )
    await userEvent.click(screen.getByRole('button', { name: /Имя/ }))
    expect(onSortChange).toHaveBeenCalledWith({ field: 'name', direction: 'desc' })
    expect(screen.getByRole('combobox', { name: 'Размер страницы' })).toBeInTheDocument()
  })

  it('does not render sort controls without a sort handler', () => {
    render(
      <DataTable
        columns={[{ ...columns[0], sortable: true }]}
        rows={[{ id: '1', name: 'Анна' }]}
        getRowId={getRowId}
      />,
    )
    expect(screen.queryByRole('button', { name: /Имя/ })).not.toBeInTheDocument()
  })

  it('shows empty state', () => {
    render(<DataTable columns={columns} rows={[]} getRowId={getRowId} emptyTitle="Пусто" />)
    expect(screen.getByText('Пусто')).toBeInTheDocument()
  })

  it('shows an error instead of the table', () => {
    render(
      <DataTable columns={columns} rows={undefined} getRowId={getRowId} error={new Error('x')} />,
    )
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })
})
