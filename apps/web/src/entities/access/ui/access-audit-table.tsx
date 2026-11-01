import type { ReactNode } from 'react'
import { Link } from 'react-router'
import { formatDateTime } from '@/shared/lib/format'
import { DataTable, type Column } from '@/shared/ui/data-table'
import { useAccessAudit } from '../api/use-access'
import { auditSummary, isGrantAuditAction } from '../model/audit-summary'
import type { AccessAuditFilters } from '../model/keys'
import { auditActionLabel } from '../model/labels'
import type { AccessAuditEntry } from '../model/types'

interface AccessAuditTableProps {
  filters: AccessAuditFilters
  actorLabel: (entry: AccessAuditEntry) => ReactNode
  onPageChange: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
  emptyTitle: string
  onResetFilters?: () => void
}

const objectCell = (entry: AccessAuditEntry) =>
  isGrantAuditAction(entry.action) ? (
    <Link to={`/access/grants/${entry.entityId}`} className="underline-offset-4 hover:underline">
      Грант {entry.entityId}
    </Link>
  ) : (
    <span className="break-all">{entry.entityId}</span>
  )

export function AccessAuditTable({
  filters,
  actorLabel,
  onPageChange,
  onPageSizeChange,
  emptyTitle,
  onResetFilters,
}: AccessAuditTableProps) {
  const audit = useAccessAudit(filters)
  const columns: Column<AccessAuditEntry>[] = [
    { id: 'createdAt', header: 'Время', cell: (entry) => formatDateTime(entry.createdAt) },
    { id: 'actor', header: 'Кто', cell: actorLabel },
    { id: 'action', header: 'Действие', cell: (entry) => auditActionLabel(entry.action) },
    { id: 'object', header: 'Объект', cell: objectCell },
    {
      id: 'details',
      header: 'Подробности',
      cell: (entry) => <span className="text-sm break-words">{auditSummary(entry) || '—'}</span>,
    },
  ]

  return (
    <DataTable
      columns={columns}
      rows={audit.data?.items}
      getRowId={(entry) => entry.id}
      isLoading={audit.isLoading}
      error={audit.error}
      emptyTitle={emptyTitle}
      onResetFilters={onResetFilters}
      meta={audit.data?.meta}
      onPageChange={onPageChange}
      onPageSizeChange={onPageSizeChange}
    />
  )
}
