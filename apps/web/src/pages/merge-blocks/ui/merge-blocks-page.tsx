import { mergeBlockReasonLabels, useMergeBlocks, type MergeBlock } from '@/entities/merge-block'
import { DeleteMergeBlockButton } from '@/features/delete-merge-block'
import { formatDateTime } from '@/shared/lib/format'
import { useListParams } from '@/shared/lib/list-params'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { PageHeader } from '@/shared/ui/page-header'

const columns: Column<MergeBlock>[] = [
  { id: 'createdAt', header: 'Создан', cell: (b) => formatDateTime(b.createdAt) },
  { id: 'guestA', header: 'Профиль A', cell: (b) => b.guestAId },
  { id: 'guestB', header: 'Профиль B', cell: (b) => b.guestBId },
  { id: 'reason', header: 'Причина', cell: (b) => mergeBlockReasonLabels[b.reason] },
  {
    id: 'actions',
    header: '',
    cell: (b) => <DeleteMergeBlockButton blockId={b.id} />,
    className: 'text-right',
  },
]

export function MergeBlocksPage() {
  const { page, setPage } = useListParams()
  const blocks = useMergeBlocks({ page })

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Запреты объединения"
        description="Пары профилей, которые не будут объединены автоматически"
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
