import { useState } from 'react'
import { rejectedCodeLabels, useRejectedRows, type RejectedRow } from '@/entities/import-batch'
import { DataTable, type Column } from '@/shared/ui/data-table'

const columns: Column<RejectedRow>[] = [
  { id: 'row', header: 'Строка', cell: (r) => r.rowNumber, className: 'w-24' },
  { id: 'code', header: 'Причина', cell: (r) => rejectedCodeLabels[r.errorCode] },
  { id: 'message', header: 'Описание', cell: (r) => r.errorMessage },
  {
    id: 'raw',
    header: 'Данные',
    cell: (r) => (r.rawData ? <code className="text-xs">{JSON.stringify(r.rawData)}</code> : '—'),
  },
]

export function RejectedTable({ batchId }: { batchId: string }) {
  const [page, setPage] = useState(1)
  const rows = useRejectedRows(batchId, page)

  return (
    <DataTable
      columns={columns}
      rows={rows.data?.items}
      getRowId={(r) => r.id}
      isLoading={rows.isLoading}
      error={rows.error}
      emptyTitle="Отклонённых строк нет"
      meta={rows.data?.meta}
      onPageChange={setPage}
    />
  )
}
