import { StatusBadge, type StatusTone } from '@/shared/ui/status-badge'
import { guestStatusLabels } from '../model/status'
import type { GuestStatus } from '../model/types'

const toneByStatus: Record<GuestStatus, StatusTone> = {
  active: 'success',
  merged: 'neutral',
  deleted: 'danger',
}

export function GuestStatusBadge({ status }: { status: GuestStatus }) {
  return <StatusBadge tone={toneByStatus[status]}>{guestStatusLabels[status]}</StatusBadge>
}
