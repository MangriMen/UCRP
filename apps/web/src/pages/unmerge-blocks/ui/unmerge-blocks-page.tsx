import { useState } from 'react'
import {
  unmergeBlockReasonLabels,
  useUnmergeBlocks,
  type UnmergeBlock,
} from '@/entities/unmerge-block'
import { DeleteUnmergeBlockButton } from '@/features/delete-unmerge-block'
import { formatDateTime } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'

const columns: Column<UnmergeBlock>[] = [
  { id: 'createdAt', header: 'Создан', cell: (b) => formatDateTime(b.createdAt) },
  { id: 'guestA', header: 'Профиль A', cell: (b) => b.guestAId },
  { id: 'guestB', header: 'Профиль B', cell: (b) => b.guestBId },
  { id: 'reason', header: 'Причина', cell: (b) => unmergeBlockReasonLabels[b.reason] },
  {
    id: 'actions',
    header: '',
    cell: (b) => <DeleteUnmergeBlockButton blockId={b.id} />,
    className: 'text-right',
  },
]

export function UnmergeBlocksPage() {
  const [page, setPage] = useState(1)
  const blocks = useUnmergeBlocks({ page })

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Запреты склейки"
        description="Пары профилей, которые не будут склеены автоматически"
      />
      <DataTable
        columns={columns}
        rows={blocks.data?.items}
        getRowId={(b) => b.id}
        isLoading={blocks.isLoading}
        error={blocks.error}
        emptyTitle="Запретов нет"
        meta={blocks.data?.meta}
        onPageChange={setPage}
      />
    </div>
  )
}
