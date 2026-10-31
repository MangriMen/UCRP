import { StatusBadge, type StatusTone } from '@/shared/ui/status-badge'
import { segmentStatusLabels } from '../model/status'
import type { SegmentStatus } from '../model/types'

const toneByStatus: Record<SegmentStatus, StatusTone> = {
  pending: 'neutral',
  computing: 'info',
  ready: 'success',
  failed: 'danger',
  archived: 'neutral',
}

export function SegmentStatusBadge({ status }: { status: SegmentStatus }) {
  return <StatusBadge tone={toneByStatus[status]}>{segmentStatusLabels[status]}</StatusBadge>
}
