import { StatusBadge, type StatusTone } from '@/shared/ui/status-badge'
import { exportStatusLabels } from '../model/status'
import type { ExportStatus } from '../model/types'

const toneByStatus: Record<ExportStatus, StatusTone> = {
  queued: 'neutral',
  building: 'info',
  ready: 'success',
  failed: 'danger',
}

export function ExportStatusBadge({ status }: { status: ExportStatus }) {
  return <StatusBadge tone={toneByStatus[status]}>{exportStatusLabels[status]}</StatusBadge>
}
