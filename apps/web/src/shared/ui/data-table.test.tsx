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
