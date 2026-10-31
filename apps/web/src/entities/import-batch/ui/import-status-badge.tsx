import { StatusBadge, type StatusTone } from '@/shared/ui/status-badge'
import { importStatusLabels } from '../model/status'
import type { ImportStatus } from '../model/types'

const toneByStatus: Record<ImportStatus, StatusTone> = {
  queued: 'neutral',
  running: 'info',
  completed: 'success',
  completed_with_errors: 'danger',
  failed: 'danger',
}

export function ImportStatusBadge({ status }: { status: ImportStatus }) {
  return <StatusBadge tone={toneByStatus[status]}>{importStatusLabels[status]}</StatusBadge>
}
